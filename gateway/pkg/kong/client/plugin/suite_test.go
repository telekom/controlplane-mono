package plugin

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client"
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Suite")
}

var _ = Describe("Plugin", func() {

	Context("Util", func() {

		It("should cast the plugin", func() {
			var plugin client.CustomPlugin
			var expectedPlugin *AclPlugin

			plugin = &AclPlugin{
				Id: "123",
			}
			Expect(As(plugin, &expectedPlugin)).To(BeTrue())
			Expect(expectedPlugin.Id).To(Equal("123"))
		})

		It("should not cast the plugin", func() {
			var plugin client.CustomPlugin
			var expectedPlugin *AclPlugin

			plugin = &JwtPlugin{
				Id: "123",
			}
			Expect(As(plugin, &expectedPlugin)).To(BeFalse())
		})
	})

	Context("Jumper", func() {

		var (
			expected = &JumperConfig{
				OAuth: map[ConsumerId]OauthCredentials{
					"123": {
						ClientId:     "client-id",
						ClientSecret: "topsecret",
						Scopes:       "scope1 scope2",
					},
				},
			}
			// This must be the base64 encoded version of the expected JumperConfig
			encodedJumperConfig = "eyJvYXV0aCI6eyIxMjMiOnsiY2xpZW50SWQiOiJjbGllbnQtaWQiLCJjbGllbnRTZWNyZXQiOiJ0b3BzZWNyZXQiLCJzY29wZXMiOiJzY29wZTEgc2NvcGUyIn19fQ=="
		)

		It("should return an empty JumperConfig", func() {

			actual := NewJumperConfig()
			Expect(actual).To(Equal(&JumperConfig{
				OAuth:     map[ConsumerId]OauthCredentials{},
				BasicAuth: map[ConsumerId]BasicAuthCredentials{},
			}))

		})

		It("should correctly base64-encode", func() {
			JumperConfig := expected
			strVal := ToBase64OrDie(JumperConfig)
			Expect(strVal).To(Equal(encodedJumperConfig))
		})

		It("should correctly base64-decode", func() {
			actual, err := FromBase64(encodedJumperConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})

	Context("Encode", func() {

		It("should correctly encode a string map", func() {
			m := New()
			m.AddKV("key1", "value1")
			m.AddKV("key2", "value2")

			encoded, err := m.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			var actual []string
			err = json.Unmarshal(encoded, &actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(ConsistOf("key1:value1", "key2:value2"))
		})

		It("should correctly decode a string map", func() {
			m := New()
			err := json.Unmarshal([]byte(`["key1:value1","key2:value2"]`), m)
			Expect(err).ToNot(HaveOccurred())

			Expect(m.items).To(HaveLen(2))
			Expect(m.items).To(HaveKey("key1"))
			Expect(m.items).To(HaveKey("key2"))
			Expect(m.Get("key1")).To(Equal("value1"))
			Expect(m.Get("key2")).To(Equal("value2"))
		})
	})

})
