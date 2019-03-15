#!/usr/bin/env python3

from github import Github
import os
from logzero import logger


def collect_operators_runner():
    """ Runner entry point """
    logger.info(f'Using github api to collect all operators')
    logger.info('For api access, reading form $USER and $PASSWORD env vars..')
    logger.info('Please make sure these are set.')
    user = os.getenv("USER")
    passwd = os.getenv("PASSWORD")
    g = Github(user, passwd)
    for repo in g.get_user().get_repos():
        print(repo.name)
