### QKK (Quick Kubernetes)

QKK is a very simple cli tool that allows users to run kubectl actions on resources while 
also searching for a specific resource name.

A common pattern is to search for some kubernetes resource, copy its name, then run that 
action via a copy paste. Now all that is combined in one tool, no more taking your hands off 
the keyboard.


#### Installation

```bash
git clone git@github.com:charliemenke/qkk.git
cd ./qkk
go build .
```

You can now use the binary by calling `./qkk <options>`. Move the binary somewhere in your path 
to call it globally.


#### Usage

The `qkk` command's options are made up of three parts:
1. `--resource` or `-r`: The resource you want to take an action on. Examples being `Pod` or `Deployment`. 
This **IS** a required field.
2. `--pattern` or `-p`: A certian string pattern you want to search the resource with. This essentially 
greps for your resource using the supllied pattern. This is **NOT** a required field.
3. `<action>`. The trailing arg should be the action and any of its arguments you want to take agaisnt your resource. Examples could 
be: `logs -f` or `edit` or `describe`.
> Note: the action does not have to be a single word, this is the place to add kubectl specific arguments 
your action

Putting this all together: `qkk -r=pod -p=staging logs -f`. Running this command would get all kubernetes 
that have `"dev"` in their names, present the list of matching pods to you to select, then run: `kubectl logs -f <your selection>`.
