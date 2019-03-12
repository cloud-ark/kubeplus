#!/usr/bin/env python3
from logzero import logger
from analysis.utils import *

class Guidelines:
    def __init__(self, repo_name):
        self.repo_name = repo_name
    def test_crd_registered_in_helm_chart(self):

        return False
    def test_owner_references_set(self):
        owner_ref_regex = b'OwnerReferences'
        has_owner_ref = search_for_key(self.repo_name,owner_ref_regex)
        return has_owner_ref

    def test_kube_openapi_annotations_on_typedefs(self):
        api_annotations_regex = b'k8s:openapi-gen=true'
        has_kube_api_annotation = search_for_key(self.repo_name,api_annotations_regex)
        return has_kube_api_annotation

    def test_has_custom_resource_validation(self):
        validation_regex = b'validation:'
        has_validation = search_for_key(self.repo_name,validation_regex)
        return has_validation

    def test_helm_chart_exists(self):
        file_name = "Chart.yaml"
        has_chart = search_for_file(self.repo_name,file_name)
        return has_chart


def resolve(args):
    if args.github:
        analyze_single()
    elif args.list:
        inpf = open(args.list, 'r')
        outf = open('results.txt', 'w')
        analyze(inpf,outf)
    return

def analyze(inpf, outf):
    for line in inpf:
        repo_git = line.strip("\n")
        repo_name = get_repo_name(repo_git)
        logger.info(f'Cloning {repo_name} . . .')
        clone(repo_git)
        run_analysis(repo_git, repo_name, outf)
        logger.info(f'Cleaning up {repo_name} . . .')
        delete(repo_name)
    inpf.close()
    outf.close()
    return

def run_analysis(repo_link, repo_name, output_file):
    t = Guidelines(repo_name=repo_name)
    logger.info(f'Operator: {repo_name}')
    crd_registered_in_helm_chart = "satisfied" if t.test_crd_registered_in_helm_chart() else "not satisfied"
    owner_references_set = "satisfied" if t.test_owner_references_set() else "not satisfied"
    kube_openapi_annotations_on_type_definitions = "satisfied" if t.test_kube_openapi_annotations_on_typedefs() else "not satisfied"
    custom_resource_spec_validation = "satisfied" if t.test_has_custom_resource_validation() else "not satisfied"
    helm_chart_exists = "satisfied" if t.test_helm_chart_exists() else "not satisfied"
    logger.info(f'\tcrd_registered_in_helm_chart: {crd_registered_in_helm_chart}')
    logger.info(f'\towner_references_set: {owner_references_set}')
    logger.info(f'\tkube_openapi_annotations_on_type_definitions: {kube_openapi_annotations_on_type_definitions}')
    logger.info(f'\tcustom_resource_spec_validation: {custom_resource_spec_validation}')
    logger.info(f'\thelm_chart_exists: {helm_chart_exists}')

    output_file.write("%s, crd_registered_in_helm_chart: %s, owner_references_set: %s, \
kube_openapi_annotations_on_type_definitions: %s, custom_resource_spec_validation: %s, \
helm_chart_exists: %s " %(repo_link, \
                    crd_registered_in_helm_chart, owner_references_set, custom_resource_spec_validation,
                    custom_resource_spec_validation, helm_chart_exists))
    return
