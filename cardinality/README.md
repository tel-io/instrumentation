# Library for replace high cardinality values

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

```go
package main

import (
   "fmt"

   "github.com/tel-io/instrumentation/cardinality"
   "github.com/tel-io/instrumentation/cardinality/auto"
   "github.com/tel-io/instrumentation/cardinality/rules"
)

func main() {
   urls := []string{
      "/player/delete/file/favicon.ico",
      "/player/update/123/550e8400-e29b-41d4-a716-446655440000",
   }

   fmt.Println("\nOriginal URL")
   for _, url := range urls {
      fmt.Println(url)
   }

   fmt.Println("\nAutomatic replacing: ID, Filename, and UUID")
   r1 := auto.New()
   for _, url := range urls {
      fmt.Println(r1.Replace(url))
   }

   fmt.Println("\nReplace by custom rules")
   r2, err := rules.New([]string{
      "/player/delete/file/:resource", //Full rule (start with separator)
      "update/:id/:uuid",              //Partial rule (without leading separator)
   })
   if err != nil {
      panic(err)
   }
   for _, url := range urls {
      fmt.Println(r2.Replace(url))
   }

   fmt.Println("\nAlso you can use list of replacers for continuous processing [/shop/my-super-store/update/123]")
   r3, err := rules.New([]string{"shop/:shop-name"})
   if err != nil {
      panic(err)
   }
   rList := cardinality.ReplacerList{r1, r3}
   fmt.Println(rList.Apply("/shop/my-super-store/update/123"))

   fmt.Println("\nAnd configuring replacers [shop.123.update.550e8400-e29b-41d4-a716-446655440000]")
   cfg := cardinality.NewConfig(
      //hasLeadingSeparator:false, separator:"."
      cardinality.WithPathSeparator(false, "."),
      //placeholderRegexp:redudancy with auto, placeholderFormatter:func(string) string
      cardinality.WithPlaceholder(nil, func(id string) string {
         return fmt.Sprintf(`{%s}`, id)
      }),
   )
   r4 := auto.New(
      auto.WithoutId(),
      auto.WithConfigReader(cfg),
   )
   fmt.Println(r4.Replace("shop.123.update.550e8400-e29b-41d4-a716-446655440000"))
}

```
