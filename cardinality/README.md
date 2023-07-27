# Library for replace cardinality params

### Possible cases:
1. **Uncontrolled sender**. In the case of a controlled frontend, it is better to change it, of course))
   But uncontrolled frontend can generate wrong URL. It is also possible to fix this at the receiver level
   For valid routing rules the middleware will process such requests and return a parsed URL with placeholders.
   But for unexpected or non-existent links, a routing error will occur. 
   The problem of high cardinality can be quickly fixed by disabling non-existent handlers, but this is probably not what you would like. 
   Because then it is difficult to track errors in the query logic


2. **Bruteforce URL**. Someone might be trying to spam your system on purpose, or they just won't worry about it happening

Therefore, there was a need to reduce requests with great cardinality in automatic or configured mode, оr combinations of them.
The following implementations are currently available:


