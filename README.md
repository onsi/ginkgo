Todo
====

- Godescribe:
    X Pending and Focused should play nice.  in particular, we should know if a focused spec is pending or not:
    X really need the whole stack trace for exceptions
    - in reporter, pull out padding stuff into a helper method that takes an indentation level, a max-width, and a string.  all reporting should go through this method to make sure we have a nice constant width report.  similarly, the dots should wrap after that width
    - rename to ginkgo
    - add support for parallel runs at the process level:
        - flags that denote node# and # of nodes
        - reporter should always include the node#?
        - add an example bash script or something that takes the output and multiplexes it correctly?
    - should see top level group *if* there's a failure there.  Call it [toplevel].  If no failure there, don't show it.
    - make timeout configurable
    - clean up the way fail is propagated

- Gomega:
    - Look into some sort of matcher a la suraci's Expect
    - Clean up matchers (lots of files + tests under matchers?)

- Other:
    - add documentation
    - add snippets for sublime?