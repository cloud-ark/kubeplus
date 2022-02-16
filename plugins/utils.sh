check_namespace() {
  local ns=$1
  local kubeconfg=$2
#  ns_output=`kubectl get ns $ns $kubeconfig 2>&1 | awk '{print $1}'`
  ns_output=`kubectl get ns $ns $kubeconfig 2>&1`
  if [[ $ns_output =~ 'Error' ]]; then
     echo "Namespace $ns not found."
     exit 0
  fi
  if [[ $ns_output =~ 'Unable' ]]; then
     echo $ns_output
     exit 0
  fi
}

check_kind() {
  local kind=$1

  canonicalKindPresent=`kubectl api-resources | grep -w $kind`
  OLDIFS=$IFS
  IFS=' '
  read -a canonicalKindPresentArr <<< "$canonicalKindPresent"
  IFS=$OLDIFS

  if [[ "${#canonicalKindPresentArr}" == 0 ]]; then
    echo "Unknown Kind $kind"
    exit 0
  fi

}
