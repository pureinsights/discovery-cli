import click
import datetime

from kubernetes import client, config
from kubernetes.client.rest import ApiException

@click.group()
@click.pass_context
def cluster(ctx):
    """
    Provides common operations for an existing pdp cluster. Depends on kubectl being properly configured on your
    environment. It will use the active kubectl context. To see the current kubectl context do 'kubectl config
    current-context'.
    """
    pass

@cluster.command('restartAll')
@click.pass_context
def restart_all(ctx):
    """
    Restarts all pods belonging to the PDP in the given namespace. Uses the pre-configured
    restart policy for each deployment. Does not restart stateful sets or daemon sets.
    """
    namespace = ctx.obj['namespace']

    config.load_kube_config()
    print(f"Will use Kubernetes context {config.list_kube_config_contexts()[1]['name']}")

    k8s_client = client.AppsV1Api()

    deployment_list = k8s_client.list_namespaced_deployment(namespace=namespace)

    if deployment_list.items:
        print(f"Restarting all PDP pods in the namespace {namespace}...")
    else:
        print(f"No PDP deployments in the namespace '{namespace}'.")

    for deployment in deployment_list.items:
        deployment_name = deployment.metadata.name

        if deployment_name.startswith('pdp-'):
            restart_deployment(k8s_client, deployment_name, namespace)
            print(f"Restarted deployment {deployment_name}")

@cluster.command('fetchImageIds')
@click.pass_context
def fetch_image_ids(ctx):
    """
     Returns the current image SHA used by each PDP pod in the given namespace. This is useful
     after a code push to make sure you have the correct version.
    """
    namespace = ctx.obj['namespace']

    k8s_client = client.CoreV1Api()

    pod_list = k8s_client.list_namespaced_pod(namespace=namespace)

    if pod_list.items:
        print(f"Found {len(pod_list.items)} pods in {namespace}...")
    else:
        print(f"No PDP pods in the namespace '{namespace}'.")

    image_ids = set()

    for pod in pod_list.items:
        pod_name = pod.metadata.name

        if pod_name.startswith('pdp-'):
            for status in pod.status.container_statuses:
                image_ids.add(status.image_id)

    print(f"Found {len(image_ids)} unique image ids")

    for pod in image_ids:
        print(f"Image ID: {pod}")


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


