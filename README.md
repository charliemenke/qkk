### QKK (Quick Kubectl)

QKK is a simple cli tool that allows users to run (most) kubectl verbs on resources while 
also searching for a specific resource name.

A common pattern is to search for some kubernetes resource, copy its name, then run that 
action via a copy paste. Now all that is combined in one tool, no more taking your hands off 
the keyboard. Instead, `qkk` will present you with a quick list of resources that matched 
your optional pattern for you to select from.


#### Installation

If you dont already have `kubectl` installed, [install it](https://kubernetes.io/docs/tasks/tools/#kubectl) 
and point it at your cluster.

Next, build the `qkk` binary.

```bash
git clone git@github.com:charliemenke/qkk.git
cd ./qkk
go build .
```

You can now use the binary by calling `./qkk <options>`. Move the binary somewhere in your path 
to call it globally.


#### Usage

```bash
qkk % ./qkk --help
Usage of qkk:
  qkk [-n NAMESPACE] -r RESOURCE [-p PATTERN] ACTION ...
Options:
  -n, --namespace NAMESPACE        Search and take action in k8s namespace NAMESPACE. default: 'default'
  -r, --resource RESOURCE          Search and take action on k8 resource RESOURCE.
  -p, --pattern PATTERN            Search k8 RESOURCE by pattern PATTERN. default: ''
  -h, --help                       prints help information
```

> Note: the ACTION does not have to be a single word, this is the place to add kubectl specific arguments 
your action

Putting this all together: `qkk -r=pod -p=staging logs -f`. Running this command would get all kubernetes pods 
in the default namespace that have `"staging"` in their names, present the list of matching pods to you to select from, 
then run: `kubectl logs -f <your selection>`.

You can navigate the quick list with arrow keys, or using vim keys `h`, `j`, `k`, `l`.
