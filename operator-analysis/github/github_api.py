#!/usr/bin/env python3

from github import Github
import os
from logzero import logger


def collect_operators_runner():
    """ Runner entry point, outputs alloperators.txt file."""
    logger.info(f'Using github api to collect all operators')
    logger.info('For api access, reading from $USER and $PASSWORD env vars..')
    logger.info('Please set these to your github user and passwd.')
    user = os.getenv("USER")
    passwd = os.getenv("PASSWORD")
    g = Github(user, passwd)
    pagination_of_repo = g.search_repositories("kubernetes+operators")
    with open("../operator-repos.txt", 'w') as outf:
        for repo in pagination_of_repo:
            clone_url = repo.clone_url
            last_commit = repo.pushed_at
            num_contributors = repo.get_contributors().totalCount
            stars = repo.stargazers_count

            outf.write("clone_url:%s, last_commit:%s, \
num_contributors:%s, \
stars:%s\n" % (clone_url,
                last_commit, num_contributors,
                stars))
