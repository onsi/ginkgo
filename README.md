Todo
====

- Godescribe:
    - build out node hierarchy
        X container_node can add each of the four types
        X runnable_node can detect function type
            X runnable_node can run the function and return:
                state, timing, failureMessage, failureCodeLocation
        X it node has extra information on it

    - convert node hierarchy into flat list of examples
        - example is inited with an it
           - then add containers working outwards
           - flag(): pending > focus > none
           - can run() or skip()
               - tells containers to run bef, then run jbef, then runs it, then run aft (backwards)
               - each of these returns:
                   pass/fail/panic, err message, err codeLocation, failedComponent codeLocation
       - example report is (modify reporter_interface.go):
           - [d,d,c,d,c,d,i] (texts)
           - [cL, cL, etc...]
           - state is pass/fail/panic/pending/skip/willRun
           - run time
           - failure object:
               - failure message ({}instance?)
               - failureContainerIndex
               - failureCodeLocation
               - failureComponentType (bef/aft/jbef/it)
               - failureComponentCodeLocation


    - randomize the examples
        - by top-level node
        - all

    - report to the reporter

    - build out the reporter (fun!)
    - make timeout configurable
    - search for and do: //todo: can we get the code location that threw the panic from within the defer



- Gomega:
    - Gomega bootstrap
    - Equals Matcher
