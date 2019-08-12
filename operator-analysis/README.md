# Operator Analysis tooling

This repository contains the tools that we are developing for analyzing Kubernetes Operators.

Current analysis results and report is available on this blog post:

https://medium.com/@cloudark/analysis-of-open-source-kubernetes-operators-f6be898f2340


# Instructions

## Run script to collect all operators on Github with metadata

1. `pip3 install -r requirements.txt`

2. `cd github`

3. `python3 main.py`

## Run analysis of operators

2. `python3 main.py ./operator-repos.txt`
