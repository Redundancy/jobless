Jobless
=======

Jobless is a silly name for an experiment in reorganizing how the entrypoints to "verbs" in a project can be organised (ie. "build", "test", "run")

There are a few ideas under test here:
* Firstly, the idea that entrypoints are best organised by being allowed to be kept next to the code for them wherever that is in the project, but exposed for discovery and use from a top level.
* That they benefit from being able to share variables based on the project structure: Things like GOPATH, or branch root folder, shared without repetition to all files below.
* That expressing relative paths consistently based on the path relative to the file are expressed in is cognitively easier.
* That you rarely want the CWD to be inherited by project entrypoints

In the end, you should not have to express "../../../" in many cases.
You should be able to give a user the ability to run all tests, from potentially multiple programming languages and multiple sub-projects, without having to make a batch script that calls down to each of them.

I use YAML for a few reasons:
* It's a *lot* more readable and shorter than JSON
* You can comment in YAML, which makes for better configuration files

So here's an example of jobless in use to build itself:

I put this file in my root workspace.

```yaml
Variables:
    GOPATH: "{{ path `.` }}"
    OUTPUT_FOLDER: "{{ path `.` }}"
```

Here is an example of using those variables further down the tree:

```yaml
Tasks:
    - Name: jobless.test
      # Automatically parsing the command and arguments from a string would be nicer
      Command: ["go", "test", "-v", "."]
      Environment:
            # This is using go's text template system for convenience
            # A more readable mechanism would be better
            GOPATH: "{{var `GOPATH`}}"  
    # Because this is YAML, you can comment
    # This command should be translated to run cmd to execute the batch script          
    - Name: jobless.batch.test1
      Command: ["test.bat"]
    # This one does not need translation
    - Name: jobless.batch.test2
      Command: ["cmd", "/c", "test.bat"]
```

At the moment, jobless doesn't look above the cwd for files, but it might be a nice enhancement.
Notice that the file is relatively readable, commented and tasks are (by definition) locked down to a single command. You don't need multiple files to define multiple entrypoints.

Now we can discover these tasks from anywhere above them in the filesystem using:
```cmd
jobless find
> jobless.build in E:\stuff\GoFiddle\src\github.com\Redundancy\jobless\cmd\jobless.jobless
> jobless.test in E:\stuff\GoFiddle\src\github.com\Redundancy\jobless\lib.jobless
> jobless.batch.test1 in E:\stuff\GoFiddle\src\github.com\Redundancy\jobless\lib.jobless
> jobless.batch.test2 in E:\stuff\GoFiddle\src\github.com\Redundancy\jobless\lib.jobless
```

and we can execute them using a few methods:

* ```jobless run **``` - run everything in sequence
* ```jobless run jobless.build``` - run just jobless.build
* ```jobless run jobless.*``` - run jobless.build and jobless.test
* ```jobless run jobless.**``` - run everything in the jobless namespace
* ```jobless run **.test``` - run everything ending in test

(you hopefully get the idea)

Note that 'find' also takes these arguments.

### Why? ###

Batch files are *terrible*.
* Sharing variables largely has to be done by setting them in other things you call
* Unset variables often require many line batch tests using awful notation
* As with go projects that you `go get`, you don't always have access to put he helpful batch scripts at the root
* It's nicer to put entrypoints next to the code that they're supposed to run
* You want to be able to find all the things someone thinks you can "do" with a project
* I've often wanted to "run all the tests" on a branch of a large project, but found it difficult to get to them all and run them
* Be able to list out, debug the values of, and check for the presence of any variables easily 

### Further ideas ###
* Tasks that run other tasks in sequence (test, then build)
* Hidden tasks (tasks that are only supposed to be run as part of a sequence)
* Encrypted variables (consolidate your settings in VCS securely)
* Overrides - set an overide for a variable in jobless
* Lookups - do a lookup for a variable value based on something else

### Go crazy ###

Once you've got things running commands in sequence, with easy configuration, how far away are you from having a much nicer way of doing a CI system?

Why do CI systems NOT get their configuration from source control? it allows for much better integration behaviour between branches,
and if you're using the same system for your entrypoints, you're already building something at a nice level of granularity that's not just for one use.
