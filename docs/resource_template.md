# Resource Template

The resource template is a definition of datacenter resource.

The basic and minimal definition is as following:

```json
{
  "$schema": "https://raw.githubusercontent.com/axsh/openvdc/master/schema/v1.json#",
  "title": "MyResource",
  "template": {
    "type": "none"
  }
}
```

Save the template as ``./templates/myresource.json`` so that you can
check the syntax.

```bash
openvdc template validate ./templates/myresource.json
```

Show details about the template.

```bash
% openvdc template show ./templates/myresource.json
Type: none
Title: MyResource

Parameters overwrite:
  param1: Test parameter1
  param2: Test parameter2
```

Overwrite the parameter.

```json
{
  "$schema": "https://raw.githubusercontent.com/axsh/openvdc/master/schema/v1.json#",
  "title": "MyResource",
  "template": {
    "type": "none",
    "param1": "xxx"
  }
}
```

```bash
openvdc template validate ./templates/myresource.json --param1=yyy
```

``template validate`` returns error if undefined parameter is given.

```bash
% openvdc template validate ./templates/myresource.json --param3=x
ERROR: No such parameter: param3
```
