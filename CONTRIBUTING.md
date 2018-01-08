In order to expedite the process of me debugging the code, please change the following line in rc.local:
```
[path/to/prometheus]/prometheus &
```
to
```
[path/to/prometheus]/prometheus > [path/to/prometheus]/pometheuslogs &
```
which will output any error messeges to the `prometheuslogs` file.

Then include the output of that file in any Issues Request.
