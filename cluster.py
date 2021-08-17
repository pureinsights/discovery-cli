import datetime
import sys

from kubernetes import client, config
from kubernetes.client.rest import ApiException


def description():
    return 'cluster: common operations for managing a PDP cluster'


def print_help():
    print("""
Provides common operations for an existing pdp cluster. Depends on kubectl being properly configured on your 
environment. It will use the active kubectl context. To see the current kubectl context do 'kubectl config 
current-context'.

Usage:
    pdp cluster [options]

    Options can be:
    *restart-all [namespace]: restarts all pods belonging to the PDP in the given namespace. Uses the pre-configured
     restart policy for each deployment. Does not restart stateful sets or daemon sets. 
    """)


def run(argv, commands, configuration):
    if len(sys.argv) < 3:
        print('''
            Missing argument, the namespace where to run. For help, type 'pdp help cluster'.
        ''')

        sys.exit()

    namespace = sys.argv[2]

    config.load_kube_config()
    k8s_client = client.AppsV1Api()

    deployment_list = k8s_client.list_namespaced_deployment(namespace=namespace)

    if deployment_list.items:
        print(f"Restarting all PDP pods in the namespace {namespace}...")
    else:
        print(f"No PDP pods in the namespace '{namespace}'.")

    for deployment in deployment_list.items:
        deployment_name = deployment.metadata.name

        if deployment_name.startswith('pdp-'):
            restart_deployment(k8s_client, deployment_name, namespace)
            print(f"Restarted deployment {deployment_name}")


def restart_deployment(v1_apps, deployment, namespace):
    now = datetime.datetime.utcnow()
    now = str(now.isoformat("T") + "Z")
    body = {
        'spec': {
            'template': {
                'metadata': {
                    'annotations': {
                        'kubectl.kubernetes.io/restartedAt': now
                    }
                }
            }
        }
    }
    try:
        v1_apps.patch_namespaced_deployment(deployment, namespace, body, pretty='true')
    except ApiException as e:
        print("Exception when calling AppsV1Api->read_namespaced_deployment_status: %s\n" % e)


