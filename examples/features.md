# Features

Here is a brief explanation of KubePlus's advanced features and how to use them.

## Application instance update

An application instance can be updated by modifying the spec property of the app instance
YAML and executing `kubectl apply` on that instance. KubePlus will update that specific application instance.


## Bulk update of all application instances

This requirement arises when there are multiple application instances that need to be updated to a new version of 
the application Helm chart. This can be achieved by updating the Helm chart in the `ResourceComposition` that
defines the Kind corresponding to all such application instances, and then executing `kubectl apply` on the
resourcecomposition. KubePlus will upgrade all the application instances to the new version of the chart.
Note that the new version of the chart should contain same values.yaml as the earlier version.

## Defining per instance resource quotas

This requirement arises when resource quota needs to be defined corresponding to an application instance.
In the ResourceComposition definition, `quota` can be defined [see this](multitenancy/application-hosting/wordpress/wordpress-service-composition-localchart.yaml).
KubePlus will add a `ResourceQuota` object when it creates namespaces corresponding to application instances.

## Licensing

A License can be defined for a `Kind`. A license can have a specific expiry date, or it can limit
the number of application instances that can be created of that Kind. A license can have both.
Only one license can be defined for a `Kind`. KubePlus kubectl plugins are provided to create, get, and delete license.
A license is stored as a configmap in the namespace where KubePlus is running.
The name of the configmap has the following syntax: `<Kind>-license`.
Whenever an application instance is being created, KubePlus checks if there is a license defined for that Kind.
If there is, KubePlus enforces the licensing requirements.


## Migrating an application instance from one Kind to another Kind

This requirement can arise when using licensing feature of KubePlus. 
Say you have defined a Kind named `Customer1TrialApp` with a license with some expiry date.
Once the expiry date is over, you want to migrate the running application instance to a Kind
that represents a perpetual license named, say `Customer1App`. This can be achieved by using
the migration feature.

In order to migrate an application instance from one Kind to another, follow these steps:
- Register the new Kind in the cluster. Make sure that values.yaml of the underlying Helm chart
  corresponding to the new Kind is same as values.yaml of the original Kind.
- In the YAML file representing the application, define the new Kind
- Keep the name of the application instance same as the original name
- Include the following annotation `kubeplus/migrate-from:`
  Set its value to the name of the current Kind of the application instance
- Execute `kubectl create` on the this application YAML

For an example, [see this](multitenancy/managed-service/appmigration/steps.txt)


## Cross-namespace Network traffic

