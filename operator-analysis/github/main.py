#!/usr/bin/env python3

from github_api import collect_operators_runner


def main():
    """
    Uses github api to generate a file
    with URL's and info about all kubernetes
    operators.
    """
    collect_operators_runner()


if __name__ == "__main__":
    """ This is executed when run from the command line """
    main()
