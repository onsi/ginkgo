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

    X randomize the examples
        X by top-level node
        X all

    X report to the reporter

    X reporter should shout about slow tests 
    X randomization not working?
    X build out the reporter (fun!)

    - really need the whole stack trace actually...
    - should see top level group *if* there's a failure there.  Call it [toplevel].  If no failure there, don't show it.
    - Pending and Focused should play nice.  in particular, we should know if a focused spec is pending or not.  that'll take some work.
    - in reporter, pull out padding stuff into a helper method that takes an indentation level, a max-width, and a string.  all reporting should go through this method to make sure we have a nice constant width report.  similarly, the dots should wrap after that width
    - make timeout configurable
    - clean up the way fail is propagated
    - add beacoup tests!

- Gomega:
    - Gomega bootstrap
    - Equals Matcher
