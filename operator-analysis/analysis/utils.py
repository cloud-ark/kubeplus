from git import Repo
import os
import shutil
import re


# TODO: add metrics check
# prometheus ||  custom-metrics

def search_for_key_in_file(repo_name, regex_key, file_name):
    for root, dirs, files in os.walk(repo_name):
        for file in files:
            with open(root+"/"+file, 'rb') as f:
                filetext = f.read()
                match = re.search(regex_key, filetext)
                if match is not None and file == file_name:
                    return True
    return False


def search_for_file(repo_name, file_name,):
    for root, dirs, files in os.walk(repo_name):
        for file in files:
            if file_name == file:
                return True
    return False


def search_for_folders_with_file(repo_name, file_name):
    folders = []
    for root, dirs, files in os.walk(repo_name):
        for file in files:
            if file_name == file:
                folders.append(root)
    return folders


def search_for_key(repo_name, regex_key, extension=None, ignore_dir=None):
    for root, dirs, files in os.walk(repo_name):
        for file in files:
            if extension is None:
                with open(root+"/"+file, 'rb') as f:
                    filetext = f.read()
                    match = re.search(regex_key, filetext)
                    if match is not None:
                        return True
            elif file.endswith(extension):
                if ignore_dir is not None and ignore_dir in root:
                    continue
                with open(root+"/"+file, 'rb') as f:
                    filetext = f.read()
                    match = re.search(regex_key, filetext)
                    if match is not None:
                        return True

    return False


def get_repo_name(git_repo_url):
    return git_repo_url.split("/")[-1].strip(".git")


def clone(git_repo_url):
    if git_repo_url == '':
        raise Exception("The github repository url is empty")
    # repo_name is the name of the cloned folder: $PWD/repo_name created
    repo_name = get_repo_name(git_repo_url)
    if os.path.exists(repo_name):
        delete(repo_name)
    cwd = os.getcwd()
    Repo.clone_from(git_repo_url, str(cwd)+"/"+repo_name)
    return


def delete(repo_name):
    shutil.rmtree(repo_name)
    return
