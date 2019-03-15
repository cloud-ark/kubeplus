#!/usr/bin/env python3

import argparse
from logzero import logger
from analysis import analysis


def main(args):
    """ Main entry point of the app """
    logger.info(f'Arguments passed: {args}')
    analysis.analyze(args.inputs)


if __name__ == "__main__":
    """ This is executed when run from the command line """
    parser = argparse.ArgumentParser()
    parser.add_argument('inputs')
    args = parser.parse_args()
    main(args)
