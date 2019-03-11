#!/usr/bin/env python3
from logzero import logger
from git import Repo
import os
import shutil
import re
class Guidelines:
    def __init__(self, repo_name):
        self.repo_name = repo_name
    def run_test4(self):
        return False
    def run_test7(self):
        return False
    def run_test11(self):
        return False
    def run_test12(self):
        return False
    def run_test14(self):
        return False



def resolve(args):
    if args.github:
        analyze_single()
    elif args.list:
        inpf = open(args.list, 'r')
        outf = open('results.txt', 'w')
        analyze_multiple(inpf,outf)
    return

def analyze_single():
    return

def analyze_multiple(inpf, outf):
    for line in inpf:
        repo = line.strip("\n")
        repo_name = get_repo_name(repo)
        logger.info(f'Cloning {repo_name} . . .')
        clone(repo)
        logger.info(f'Analyzing . . .')
        # run_analysis(repo_name, outf)
        search_for_key(repo_name,b'validation:')
        logger.info(f'Finished Analysis . . .')
        logger.info(f'Cleaning up {repo_name} . . .')
        delete(repo_name)
    inpf.close()
    outf.close()
    return

def get_repo_name(repo):
    return repo.split("/")[-1].strip(".git")

def run_analysis(repo_name, output_file):
    output_file.write(f'Operator: {repo_name}\n')
    t = Guidelines(repo_name=repo_name)
    output_file.write(f'\t4. {"Satisfied" if t.run_test4() else "Not Satisfied"}\n')
    output_file.write(f'\t7. {"Satisfied" if t.run_test7() else "Not Satisfied"}\n')
    output_file.write(f'\t11. {"Satisfied" if t.run_test11() else "Not Satisfied"}\n')
    output_file.write(f'\t12. {"Satisfied" if t.run_test12() else "Not Satisfied"}\n')
    output_file.write(f'\t14. {"Satisfied" if t.run_test14() else "Not Satisfied"}\n')
    return

def search_for_key(repo_name, regex):
    match=False
    for root, dirs, files in os.walk(repo_name):
        for file in files:
            f = open(root+"/"+file, 'rb')
            filetext = f.read()
            match = re.search(regex, filetext)
            if match is not None:
                match=True
            f.close()
    return match

def clone(git_repo):
    cwd = os.getcwd()
    repo_name = get_repo_name(git_repo)
    Repo.clone_from(git_repo, str(cwd)+"/"+repo_name)
    return

def delete(repo_name):
    shutil.rmtree(repo_name)
    return
