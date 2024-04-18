package spcontext

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/onsi/gomega"
)

func TestFieldsWith(t *testing.T) {
	g := goblin.Goblin(t)
	gomega.RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Happy path", func() {
		f := &Fields{}
		newFields := f.With("test", 1, "test2", 2)
		g.It("empty base fields", func() {
			gomega.Expect(f.EvaluateFields()).To(gomega.Equal([]any(nil)))
		})
		g.It("non empty new fields", func() {
			gomega.Expect(newFields.EvaluateFields()).To(gomega.Equal([]any{"test", 1, "test2", 2}))
		})
	})
	g.Describe("odd number of fields", func() {
		f := &Fields{}
		g.It("odd number of fields", func() {
			gomega.Expect(func() {
				f.With("test", 1, "test2") //nolint:staticcheck // negative test case
			}).To(gomega.PanicWith("invalid Fields.With call: odd number of arguments"))
		})
	})
	g.Describe("key is not string", func() {
		f := &Fields{}
		g.It("odd number of fields", func() {
			gomega.Expect(func() {
				f.With("test", 1, 3, 2)
			}).To(gomega.PanicWith("invalid Fields.With call: non-string log field key: 3"))
		})
	})

}
