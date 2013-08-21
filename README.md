Todo
====

- Godescribe:
    - build out node hierarchy
        X container_node can add each of the four types
        X runnable_node can detect function type
            X runnable_node can run the function and return:
                state, timing, failureMessage, failureCodeLocation
        X it node has extra information on it

        X add code to create examples in the containers

    X convert node hierarchy into flat list of examples
        X example is inited with an it
           X then add containers working outwards
           X flag: pending > focus > none
           X can run() or skip()
               X tells containers to run bef, then run jbef, then runs it, then run aft (backwards)
               X each of these returns:
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

    X randomize the examples
        - by top-level node
        - all

    X report to the reporter

    - clean up the way fail is propagated

    X reporter should shout about slow tests 
    X randomization not working?
    X build out the reporter (fun!)
    - make timeout configurable
    - search for and do: //todo: can we get the code location that threw the panic from within the defer



- Gomega:
    - Gomega bootstrap
    - Equals Matcher
