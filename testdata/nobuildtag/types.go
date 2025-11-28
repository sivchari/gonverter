package nobuildtag

import "github.com/sivchari/gonverter/runtime"

type A struct{ X int }
type B struct{ X int }

// No build tag (should not be detected)
var _ = runtime.Register[*A, *B]()
