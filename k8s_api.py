from __future__ import print_function
import kubernetes.client
from pprint import pprint

# Configure API key authorization: BearerToken
token = 'eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImFwaXVzZXItdG9rZW4tazJ0bDIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiYXBpdXNlciIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImVjYWM0NGE3LTExODEtMTFlYS05MGJkLTUyNTQwMDIzNzZlZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFwaXVzZXIifQ.IpqNMFi6FAC049zlfhiUQ_rgm3akejIBkJeo1OLgnin6YjKidl6S3y3A7ak9v1SibZ-CkrFWwJu0dtRCrolUFVutC9-W0JQOAI5cQIw1dS6Zv1EjInk4x8tp5rjYti1l_o2rorsuw7RPcf4rTXsQ3PAmABxFb-QeSpDtMXbEeIN9qbkX_h99G__Nv1JtZgtZnDmgbS6ecTxBq3D9tm7EW_v_cxlAAz0NGC8G9zAe1W-NAEMI6mH_DQSDlU0o5HSuxIXfzXPhlo-UtnelJKpx5dKy0KDhEufQLAbbCBZD1d_sjtCk5Z2eShrxl7soNTaHia6NgS9mYuYbipwYTkY36Q'

apiserver = "https://192.168.66.155:6443"

configuration = kubernetes.client.Configuration()

configuration.host = apiserver

configuration.verify_ssl = False

configuration.api_key = {"authorization": "Bearer " + token}

kubernetes.client.Configuration.set_default(configuration)

api_instance = kubernetes.client.AppsV1beta2Api(kubernetes.client.ApiClient(configuration))

namespace = 'default'
pretty = True

# 创建deployment
'''
body = {
            "apiVersion": "apps/v1beta2",
            "kind": "Deployment",
            "metadata": {
                "annotations": {
                    "deployment.kubernetes.io/revision": "1"
                },
                "creationTimestamp": "2019-11-22T05:55:14Z",
                "generation": 1,
                "labels": {
                    "app.kubernetes.io/instance": "test",
                    "app.kubernetes.io/managed-by": "Helm",
                    "app.kubernetes.io/name": "testchart",
                    "app.kubernetes.io/version": "1.16.0",
                    "helm.sh/chart": "testchart-0.1.0"
                },
                "name": "test-testchart1",
                "namespace": "default",
            },
            "spec": {
                "progressDeadlineSeconds": 600,
                "replicas": 1,
                "revisionHistoryLimit": 10,
                "selector": {
                    "matchLabels": {
                        "app.kubernetes.io/instance": "test",
                        "app.kubernetes.io/name": "testchart"
                    }
                },
                "strategy": {
                    "rollingUpdate": {
                        "maxSurge": "25%",
                        "maxUnavailable": "25%"
                    },
                    "type": "RollingUpdate"
                },
                "template": {
                    "metadata": {
                        "creationTimestamp": None,
                        "labels": {
                            "app.kubernetes.io/instance": "test",
                            "app.kubernetes.io/name": "testchart"
                        }
                    },
                    "spec": {
                        "containers": [
                            {
                                "image": "nginx:1.16.0",
                                "imagePullPolicy": "IfNotPresent",
                                "livenessProbe": {
                                    "failureThreshold": 3,
                                    "httpGet": {
                                        "path": "/",
                                        "port": "http",
                                        "scheme": "HTTP"
                                    },
                                    "periodSeconds": 10,
                                    "successThreshold": 1,
                                    "timeoutSeconds": 1
                                },
                                "name": "testchart",
                                "ports": [
                                    {
                                        "containerPort": 80,
                                        "name": "http",
                                        "protocol": "TCP"
                                    }
                                ],
                                "readinessProbe": {
                                    "failureThreshold": 3,
                                    "httpGet": {
                                        "path": "/",
                                        "port": "http",
                                        "scheme": "HTTP"
                                    },
                                    "periodSeconds": 10,
                                    "successThreshold": 1,
                                    "timeoutSeconds": 1
                                },
                            }
                        ],
                        "dnsPolicy": "ClusterFirst",
                        "restartPolicy": "Always",
                        "serviceAccount": "test-testchart",
                        "serviceAccountName": "test-testchart",
                        "terminationGracePeriodSeconds": 30
                    }
                }
            }
}


api_response = api_instance.create_namespaced_deployment(namespace, body, pretty=pretty)
pprint(api_response)
'''

# 删除deployment
'''
api_response = api_instance.delete_namespaced_deployment(name='test-testchart1', namespace=namespace,
                                                         pretty=pretty, grace_period_seconds=30)
pprint(api_response)
'''


api_instance = kubernetes.client.CoreV1Api(kubernetes.client.ApiClient(configuration))

## 创建service
body = {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
        "labels": {
            "app.kubernetes.io/instance": "test",
            "app.kubernetes.io/managed-by": "Helm",
            "app.kubernetes.io/name": "testchart",
            "app.kubernetes.io/version": "1.16.0",
            "helm.sh/chart": "testchart-0.1.0"
        },
        "name": "test-testchart1",
        "namespace": "default",
    },
    "spec": {
        "ports": [
            {
                "name": "http",
                "port": 80,
                "targetPort": 80
            }
        ],
        "selector": {
            "app.kubernetes.io/instance": "test",
            "app.kubernetes.io/name": "testchart"
        },
        "type": "NodePort"
    }
}
api_response = api_instance.create_namespaced_service(namespace, body=body, pretty=pretty)
print('api_response')


# 删除service
'''
api_response = api_instance.delete_namespaced_service("test-testchart1", namespace, pretty=pretty, grace_period_seconds=30)
pprint(api_response)
'''

api_response = api_instance.list_namespaced_service(namespace)
pprint(api_response.items)
