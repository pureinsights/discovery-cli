package buckets

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestNewDumpCommand tests the NewDumpCommand() function's error cases.
func TestNewDumpCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		url       bool
		apiKey    string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Error case
		{
			name:      "No URL",
			args:      []string{"my-bucket"},
			outGolden: "NewDumpCommand_Out_NoURL",
			errGolden: "NewDumpCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDumpCommand_Out_NameDoesNotExist",
			errGolden: "NewDumpCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/content/my-bucket/scroll": {
					StatusCode:  http.StatusNotFound,
					ContentType: "application/json",
					Body: `{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/content/my-bucket/scroll", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
			}, "Could not scroll the bucket with name \"my-bucket\"."),
		},
		{
			name:      "Sent max flag is < 1",
			args:      []string{"my-bucket", "--max", "-1"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDumpCommand_Out_InvalidMax",
			errGolden: "NewDumpCommand_Err_InvalidMax",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_InvalidMax"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_InvalidMax"),
			responses: map[string]testutils.MockResponse{},
			err:       cli.NewError(cli.ErrorExitCode, "The size flag can only be greater than or equal to 1."),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

			defer srv.Close()

			in := strings.NewReader("")
			out := &bytes.Buffer{}

			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "pretty-json")
			if tc.url {
				vpr.Set("default.staging_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.staging_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			dumpCmd := NewDumpCommand(d)

			dumpCmd.SetIn(ios.In)
			dumpCmd.SetOut(ios.Out)
			dumpCmd.SetErr(ios.Err)

			dumpCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			dumpCmd.SetArgs(tc.args)

			err := dumpCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewDumpCommand_WorkingCase tests the Dump command with a working scroll.
func TestNewDumpCommand_WorkingCase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/content/my-bucket/scroll", r.URL.Path)
		token := r.URL.Query().Get("token")
		assert.Equal(t, "3", r.URL.Query().Get("size"))
		w.Header().Set("Content-Type", "application/json")
		switch token {
		case "694eb7f378aedc7a163da908":
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(`[]`))
		case "694eb7f378aedc7a163da907":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da908",
			"content": [
                  {
                          "id": "4",
                          "creationTimestamp": "2025-12-26T16:28:59Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:59Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "4e7c8a47efd829ef7f710d64da661786",
                                  "link": "https://pureinsights.com/blog/2024/kmworld-2024-key-takeaways-from-the-exhibit-hall/",
                                  "author": "Graham Gillen",
                                  "header": "KMWorld 2024: Key Takeaways from the Exhibit Hall - Pureinsights: Key insights from KMWorld 2024: AI's impact on knowledge management, standout vendors, and challenges for traditional players adapting to AI."
                          },
                          "transaction": "694eb7cb78aedc7a163da902"
                  },
                  {
                          "id": "5",
                          "creationTimestamp": "2025-12-26T16:29:05Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:05Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "b1e3e4f42c0818b1580e306eb776d4a1",
                                  "link": "https://pureinsights.com/blog/2024/google-unveils-ai-enhanced-search-features-at-2024-io-conference/",
                                  "author": "Martin Bayton",
                                  "header": "Google Unveils AI-Enhanced Search Features at I/O Conference - Pureinsights: Google I/O 2024 Developer Conference key takeaways, including AI-generated summaries and other features for search."
                          },
                          "transaction": "694eb7d178aedc7a163da903"
                  },
                  {
                          "id": "6",
                          "creationTimestamp": "2025-12-26T16:29:12Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:12Z",
                          "action": "STORE",
                          "checksum": "228cc56c873a457041454280c448b4e3",
                          "content": {
                                  "_id": "232638a332048c4cb159f8cf6636507f",
                                  "link": "https://pureinsights.com/blog/2025/7-tech-trends-in-ai-and-search-for-2025/",
                                  "author": "Phil Lewis",
                                  "header": "7 Tech Trends in AI and Search for 2025 - Pureinsights: 7 Tech Trends is AI and Search for 2025 - presented by Pureinsights CTO, Phil Lewis. A blog about key trends to look for in the coming year."
                          },
                          "transaction": "694eb7d878aedc7a163da904"
                  }
          ],
		  "empty": false
			}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da907",
			"content": [
                  {
                          "id": "1",
                          "creationTimestamp": "2025-12-26T16:28:38Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:38Z",
                          "action": "STORE",
                          "checksum": "58b3d1b06729f1491373b97fd8287ae1",
                          "content": {
                                  "_id": "5625c64483bef0d48e9ad91aca9b2f94",
                                  "link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
                                  "author": "Graham Gillen",
                                  "header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
                          },
                          "transaction": "694eb7b678aedc7a163da8ff"
                  },
                  {
                          "id": "2",
                          "creationTimestamp": "2025-12-26T16:28:46Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:46Z",
                          "action": "STORE",
                          "checksum": "b76292db9fd1c7aef145512dce131f4d",
                          "content": {
                                  "_id": "768b0a3bcee501dc624484ba8a0d7f6d",
                                  "link": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
                                  "author": "Matt Willsmore",
                                  "header": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications."
                          },
                          "transaction": "694eb7be78aedc7a163da900"
                  },
                  {
                          "id": "3",
                          "creationTimestamp": "2025-12-26T16:28:54Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:54Z",
                          "action": "STORE",
                          "checksum": "cbffeeba8f4739650ae048fb382c8870",
                          "content": {
                                  "_id": "d758c733466967ea6f13b20bcbfcebb5",
                                  "link": "https://pureinsights.com/blog/2024/modernizing-search-with-generative-ai/",
                                  "author": "Martin Bayton",
                                  "header": "Modernizing Search with Generative AI - Pureinsights: Blog: why you should implement Retrieval-Augmented Generation (RAG) and how platforms like Pureinsights Discovery streamline the process."
                          },
                          "transaction": "694eb7c678aedc7a163da901"
                  }
          ],
		  "empty": false
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "json")
	vpr.Set("default.staging_url", srv.URL)

	vpr.Set("default.staging_key", "")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	dumpCmd := NewDumpCommand(d)

	dumpCmd.SetIn(ios.In)
	dumpCmd.SetOut(ios.Out)
	dumpCmd.SetErr(ios.Err)

	dumpCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	dumpCmd.SetArgs([]string{"my-bucket", "--filter", `{
	"equals": {
		"field": "author",
		"value": "Martin Bayton",
		"normalize": true
	}
}`, "--max", "3", "--output-file", filepath.Join(t.TempDir(), "my-bucket.zip")})

	err := dumpCmd.Execute()
	require.NoError(t, err)
	testutils.CompareBytes(t, "NewDumpCommand_Out_WorkingScroll", testutils.Read(t, "NewDumpCommand_Out_WorkingScroll"), out.Bytes())
}

// TestNewDumpCommand_NoProfileFlag tests the NewDumpCommand when the profile flag was not defined.
func TestNewDumpCommand_NoProfileFlag(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "pretty-json")

	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	dumpCmd := NewDumpCommand(d)

	dumpCmd.SetIn(ios.In)
	dumpCmd.SetOut(ios.Out)
	dumpCmd.SetErr(ios.Err)

	dumpCmd.SetArgs([]string{"my-bucket"})

	err := dumpCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDumpCommand_Out_NoProfile", testutils.Read(t, "NewDumpCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewDumpCommand_Err_NoProfile", testutils.Read(t, "NewDumpCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewDumpCommand_NotExactly1Arg tests the NewDumpCommand function when it does not receive exactly one argument.
func TestNewDumpCommand_NotExactly1Arg(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "pretty-json")

	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	dumpCmd := NewDumpCommand(d)

	dumpCmd.SetIn(ios.In)
	dumpCmd.SetOut(ios.Out)
	dumpCmd.SetErr(ios.Err)

	dumpCmd.SetArgs([]string{})

	err := dumpCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewDumpCommand_Out_NotExactly1Arg", testutils.Read(t, "NewDumpCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewDumpCommand_Err_NotExactly1Arg", testutils.Read(t, "NewDumpCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}
