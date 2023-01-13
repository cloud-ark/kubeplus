check_namespace() {
  local ns=$1
  local kubeconfg=$2
#  ns_output=`kubectl get ns $ns $kubeconfig 2>&1 | awk '{print $1}'`
  ns_output=`kubectl get ns $ns $kubeconfg 2>&1`
  if [[ $ns_output =~ 'Error' ]]; then
     echo "Namespace $ns not found."
     exit 0
  fi
  if [[ $ns_output =~ 'error' ]]; then
     echo $ns_output
     exit 0
  fi
  if [[ $ns_output =~ 'Unable' ]]; then
     echo $ns_output
     exit 0
  fi
}

check_kind() {
  local kind=$1
  local kubeconfg=$2

  canonicalKindPresent=`kubectl api-resources --kubeconfig=$kubeconfg | grep -w $kind`
  OLDIFS=$IFS
  IFS=' '
  read -a canonicalKindPresentArr <<< "$canonicalKindPresent"
  IFS=$OLDIFS

  if [[ "${#canonicalKindPresentArr}" == 0 ]]; then
    echo "Unknown Kind $kind"
    exit 0
  fi

}

get_canonical_kind() {
  local kind=$1

  #canonicalKindPresent=`kubectl api-resources $kubeconfig | grep -w $kind | awk '{print $4}'`
  canonicalKindPresent=`kubectl api-resources $kubeconfig | grep -w $kind`
  OLDIFS=$IFS
  IFS=' '
  read -a canonicalKindPresentArr <<< "$canonicalKindPresent"
  IFS=$OLDIFS

  if [[ "${#canonicalKindPresentArr}" == 0 ]]; then
    echo "Unknown Kind $kind."
    exit 0
  fi

  # For Kinds in empty API group (like Pods)
  a="${canonicalKindPresentArr[2]}" 
  if [[ $a == 'true' ]]; then
    canonicalKind="${canonicalKindPresentArr[3]}"
  fi

  # For Kinds in non-empty API group (like Ingress)
  b="${canonicalKindPresentArr[3]}" 
  if [[ $b == 'true' ]]; then
    canonicalKind="${canonicalKindPresentArr[4]}"
  fi

  echo $canonicalKind
}
