Todo
====

- Godescribe:
    - build out node hierarchy
        - container_node can add each of the four types
        - runnable_node can detect function type
            - runnable_node can run the function and return:
                state, timing, failureMessage
        - example node has extra information on it

    - convert node hierarchy into flat list of examples
        - container node can generate an array of examples
        - call it "example" -- an array of container nodes, runnable nodes, and one it node in the correct order
        - example has an ExampleState
        - example can generate an ExampleSummary and populate it with ExampleComponents
        - examples can be run
            - runs each component and builds/updates the exampleComponent structure
            - handles failures and aborts when one is found
            - returns an ExampleSummary

    - randomize the examples
        - by top-level node
        - all

    - report to the reporter

    - build out the reporter (fun!)


- Gomega:
    - Gomega bootstrap
    - Equals Matcher
