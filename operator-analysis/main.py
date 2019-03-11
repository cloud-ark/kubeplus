#!/usr/bin/env python3

__author__ = "Daniel Moore"
__version__ = "0.1.0"
__license__ = "MIT"

import argparse
from logzero import logger

from analysis1 import analysis

def main(args):
    """ Main entry point of the app """
    logger.info(args)
    analysis.resolve(args)


if __name__ == "__main__":
    """ This is executed when run from the command line """
    parser = argparse.ArgumentParser()

    # Required positional argument
    # parser.add_argument("arg", help="Required positional argument")

    # Optional argument flag which defaults to False
    parser.add_argument("-l", "--list")

    # Optional argument which requires a parameter (eg. -d test)
    parser.add_argument("-g", "--github")

    # Optional verbosity counter (eg. -v, -vv, -vvv, etc.)
    parser.add_argument(
        "-v",
        "--verbose",
        action="count",
        default=0,
        help="Verbosity (-v, -vv, etc)")

    args = parser.parse_args()
    main(args)
