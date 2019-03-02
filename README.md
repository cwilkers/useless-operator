## The Useless Operator

This project is based off the
[Operator Framework SDK](http://github.com/operator-framework/operator-sdk/)
and the example Memcached operator detailed in the
[User Guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md).

The useless operator mimics the [useless machine](https://en.wikipedia.org/wiki/Useless_machine) as
an operator whose sole function is to scale itself down to zero, in effect switching itself off.

The main loop of the operator, its Reconcile function, manages a Deployment of
Busybox Pods, and keeps the Deployment in sync with any Useless custom
resources (CRs) it finds in its namespace. In order to see the results of the
reconcile function, the steps it takes are, in order:

  - Ensure a Deployment exists corresponding to the Useless CR
  - Ensure the Deployment `replicas` is set to the Useless CR's `size` parameter
  - Ensure the Pods belonging to the Deployment are listed in the status of the Useless CR
  - Ensure the Useless CR `size` is set to 0

Creating the Useless CR with a size other than zero will scale up the Deployment, logs its pod names,
then set the size of the Useless CR to 0, and scale down the Deployment to no pods. Any attempt to edit
the size parameter of the CR will cause it to scale up and down again.
