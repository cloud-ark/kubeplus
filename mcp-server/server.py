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
def get_crd_metrics(crd_name: str, instance_name: str) -> str:
    """
    Collects metrics for a given KubePlus CRD instance using the
    kubectl metrics plugin.

    Args:
        crd_name: The CRD kind (e.g. 'mysql')
        instance_name: The specific instance name of that CRD
    """
    if not os.path.isfile(KUBECONFIG_PATH):
        return f"Server misconfiguration: kubeconfig not found at {KUBECONFIG_PATH}"

    try:
        result = subprocess.run(
            ["kubectl", "metrics", crd_name, instance_name, "-k", KUBECONFIG_PATH],
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

#if __name__ == "__main__":
#    mcp.run(transport="streamable-http")
app = mcp.streamable_http_app()

app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["*"]
)
