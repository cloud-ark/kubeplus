PlatformWorkflow Operator Need and Semantics
----------------------------------------------

In creating Kubernetes Workflows involving Custom Resources there is often a need
to create relationships based on annotation, label or spec property across Custom or built-in resources.

PlatformWorkflow CRD is designed to ease defining such workflows.

It should be applied after creating the resources involved in a workflow.
It provides mechanism to apply policies (e.g.: addlabel, addannotation) across all the
resources defined in the workflows. This involves top-level Custom Resources as well as
their sub-resources.
Currently supported policies include: addlabel, addannotation. 
Deletes: coming later (deletelabel, deleteannotation).
PlatformWorkflow also enables modeling dependency relationship. This is crucial for 
ensuring robustness of workflows, example preventing deletion of a Custom Resource that
is still in use (To showcase dependency modeling, use Multus <-> Pod example).

Q. Why defining addlabel and addannotation in PlatformWorkflow CRD is better than in any
of the Custom Resources themselves?
A. Custom Resources are not changed ever.

Q.  
