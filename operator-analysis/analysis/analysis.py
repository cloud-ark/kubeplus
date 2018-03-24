#!/usr/bin/env python3
from logzero import logger
from analysis.utils import clone, search_for_key, search_for_file, \
    get_repo_name, delete, search_for_folders_with_file
import os
import traceback
import re


class Guidelines:
    def __init__(self, repo_name):
        self.repo_name = repo_name

    def test_crd_registered_in_helm_chart(self):
        if not _has_helm(self.repo_name):
            return False

        helm_dirs = search_for_folders_with_file(self.repo_name, "Chart.yaml")
        customresource = re.compile(b'kind: CustomResourceDefinition')
        for dir in helm_dirs:
            if search_for_key(dir, customresource, extension=".yaml",
                              ignore_dir="vendor"):
                return True
        return False

    def test_owner_references_set(self):
        owner_ref_regex = re.compile(b'OwnerReferences?', re.IGNORECASE)
        has_owner_ref = search_for_key(self.repo_name, owner_ref_regex,
                                       extension=".go", ignore_dir="vendor")
        return has_owner_ref

    def test_kube_openapi_annotations_on_typedefs(self):
        api_annotations_regex = re.compile(b'// \\+k8s:openapi-gen=true')
        has_kube_api_annotation = search_for_key(self.repo_name,
                                                 api_annotations_regex,
                                                 extension=".go",
                                                 ignore_dir="vendor")
        return has_kube_api_annotation

    def test_has_custom_resource_validation(self):
        validation_regex = re.compile(b'validation:')
        has_validation = search_for_key(self.repo_name, validation_regex,
                                        extension=".yaml", ignore_dir="vendor")
        return has_validation

    def test_helm_chart_exists(self):
        return _has_helm(self.repo_name)


def _has_helm(repo_name):
    """Helper method to check whether a operator
    repository has helm set up
    """
    helm_dirs = search_for_folders_with_file(repo_name, "Chart.yaml")
    if not helm_dirs:
        return False

    for dir in helm_dirs:
        if os.path.isdir(os.getcwd() + "/" + dir + "/templates"):
            return True
    return False


def analyze(inputs_file):
    """Main loop that analyzes each line in inputs_file"""
    inpf = open(inputs_file, 'r')
    outf = open('results.txt', 'w')
    for line in inpf:
        repo_git = None
        # loading the output format of github/main.py, alloperators.txt
        repo_git = line.split(",")[0].replace("clone_url:", "")
        repo_name = get_repo_name(repo_git)
        try:
            clone(repo_git)
            run_analysis(repo_git, repo_name, outf)
            delete(repo_name)
        except Exception as e:
            traceback.print_exc()
            logger.error(e)
            if os.path.exists(repo_name):
                delete(repo_name)
            inpf.close()
            outf.close()
            return

    inpf.close()
    outf.close()
    return


def run_analysis(repo_link, repo_name, output_file):
    """
    Runs the analysis for a single repository,
    calling each guideline test. This assumes it was cloned
    into a folder called repo_name, which is used mutually by
    clone and this method.
    """
    t = Guidelines(repo_name=repo_name)
    logger.info(f'Operator: {repo_name}')

    crd_registered_in_helm_chart = "satisfied" if \
        t.test_crd_registered_in_helm_chart() else "not satisfied"

    owner_references_set = "satisfied" if t.test_owner_references_set() else \
        "not satisfied"

    kube_openapi_annotations_on_type_definitions = "satisfied" if \
        t.test_kube_openapi_annotations_on_typedefs() else "not satisfied"

    custom_resource_spec_validation = "satisfied" if \
        t.test_has_custom_resource_validation() else "not satisfied"

    helm_chart_exists = "satisfied" if t.test_helm_chart_exists() \
        else "not satisfied"

    logger.info(f'\tcrd_registered_in_helm_chart: \
{crd_registered_in_helm_chart}')
    logger.info(f'\towner_references_set: {owner_references_set}')
    logger.info(f'\tkube_openapi_annotations_on_type_definitions: \
{kube_openapi_annotations_on_type_definitions}')
    logger.info(f'\tcustom_resource_spec_validation: \
{custom_resource_spec_validation}')
    logger.info(f'\thelm_chart_exists: {helm_chart_exists}')

    output_file.write("%s, crd_registered_in_helm_chart: %s, owner_references_set: %s, \
kube_openapi_annotations_on_type_definitions: %s, \
custom_resource_spec_validation: %s, \
helm_chart_exists: %s \n" % (repo_link,
        crd_registered_in_helm_chart, owner_references_set,
        kube_openapi_annotations_on_type_definitions,
        custom_resource_spec_validation, helm_chart_exists))
    return
