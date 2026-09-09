import os
import subprocess
from mcp.server.fastmcp import FastMCP
from starlette.middleware.trustedhost import TrustedHostMiddleware
from mcp.server.transport_security import TransportSecuritySettings

mcp = FastMCP(name="KubePlusMCPServer")

mcp.settings.transport_security.allowed_hosts = [
    "*"
    "kubeplus-mcp.kagent",
    "kubeplus-mcp.kagent:8000",
    "kubeplus-mcp.kagent.svc.cluster.local",
    "kubeplus-mcp.kagent.svc.cluster.local:8000",
]

mcp.settings.transport_security.enable_dns_rebinding_protection = False

KUBECONFIG_PATH = os.environ.get(
    "KUBEPLUS_KUBECONFIG_PATH", "/etc/kubeplus/kubeconfig/config"
)

@mcp.tool()
def get_kubeplus_kubectl_commands() -> str:
    """

    Get the complete list of supported `kubectl kubeplus` commands.

    Use this tool whenever the user asks:
    - what kubeplus commands are available
    - list kubeplus commands
    - kubectl kubeplus help
    - kubeplus CLI
    - kubeplus command syntax

    This tool is the authoritative source for KubePlus CLI commands.
    Do not answer from memory.
    """

    try:
        result = subprocess.run(
            ["kubectl", "kubeplus", "commands"],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl kubeplus commands: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl kubeplus command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"
    

@mcp.tool()
def get_crd_metrics(kind: str, instance_name: str) -> str:
    """
    Collects metrics for a given KubePlus CRD instance using the
    kubectl metrics plugin.

    Args:
        kind: The Kubernetes Kind of the Custom Resource (e.g. 'MySQL')
        instance_name: The specific instance name of that CRD
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "metrics", kind, instance_name, "-k", KUBECONFIG_PATH],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl metrics: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl metrics command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def get_crd_info(kind: str) -> str:
    """
    Retrieves detailed information about a specific KubePlus CRD kind.

    Args:
        kind: The Kubernetes Kind of the Custom Resource (e.g. 'mysql')
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "man", kind, "-k", KUBECONFIG_PATH],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl man: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl man command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def list_instances(kind: str, namespace: str = "") -> str:
    """
    Lists all instances of a given KubePlus CRD kind.

    Args:
        kind: The Kubernetes Kind of the Custom Resource (e.g. 'mysql')
        namespace: Optional Kubernetes namespace. If omitted, instances
        are listed across all namespaces.
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        cmd = ["kubectl", "get", kind, "-o", "json", "--kubeconfig", KUBECONFIG_PATH]
        if namespace:
            cmd.extend(["-n", namespace])
        else:
            cmd.append("-A")

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl get: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl get command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def describe_instance(kind: str, instance_name: str, namespace: str) -> str:
    """
    Get detailed information about a KubePlus CRD instance.

    Args:
        kind: The Kubernetes Kind of the Custom Resource (e.g. 'mysql')
        instance_name: The specific instance name of that CRD
        namespace: The namespace where the instance resides
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "describe", kind, instance_name, "-n", namespace, "--kubeconfig", KUBECONFIG_PATH],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl describe: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl describe command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def get_application_resources(kind: str, instance_name: str) -> str:
    """
    Get the Kubernetes resources created for a KubePlus application instance.

    Args:
        kind: The Kubernetes Kind of the application instance.
        instance_name: The name of the application instance.
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "appresources", kind, instance_name, "-k", KUBECONFIG_PATH],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl appresources: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl appresources command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def get_related_resources(kind: str, instance_name: str, namespace: str) -> str:
    """
    Get resources related to a KubePlus application instance.

    Args:
        kind: The Kubernetes Kind of the application instance.
        instance_name: The name of the application instance.
        namespace: The namespace containing the application instance.
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "connections", kind, instance_name, namespace, "-k", KUBECONFIG_PATH, "-o", "json"],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl connections: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl connections command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"


@mcp.tool()
def get_license_status(kind: str) -> str:
    """
    Get license information for a KubePlus Custom Resource Kind.

    Args:
        kind: The Kubernetes Kind of the Custom Resource.
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "license", "get", kind, "-k", KUBECONFIG_PATH],
            capture_output=True,
            text=True,
            timeout=60,
        )
        if result.returncode != 0:
            return f"Error running kubectl license get: {result.stderr}"
        return result.stdout
    except subprocess.TimeoutExpired:
        return "kubectl license get command timed out after 60s"
    except Exception as e:
        return f"Unexpected error: {str(e)}"

#if __name__ == "__main__":
#    mcp.run(transport="streamable-http")
app = mcp.streamable_http_app()

app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["*"]
)
