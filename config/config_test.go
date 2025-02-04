package config

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/0xERR0R/blocky/helpertest"

	. "github.com/0xERR0R/blocky/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Creation of Config", func() {
		When("Test config file will be parsed", func() {
			It("should return a valid config struct", func() {
				err := os.Chdir("../testdata")
				Expect(err).Should(Succeed())

				LoadConfig("config.yml", true)

				Expect(config.DNSPorts).Should(Equal(ListenConfig{"55553", ":55554", "[::1]:55555"}))
				Expect(config.Upstream.ExternalResolvers["default"]).Should(HaveLen(3))
				Expect(config.Upstream.ExternalResolvers["default"][0].Host).Should(Equal("8.8.8.8"))
				Expect(config.Upstream.ExternalResolvers["default"][1].Host).Should(Equal("8.8.4.4"))
				Expect(config.Upstream.ExternalResolvers["default"][2].Host).Should(Equal("1.1.1.1"))
				Expect(config.CustomDNS.Mapping.HostIPs).Should(HaveLen(2))
				Expect(config.CustomDNS.Mapping.HostIPs["my.duckdns.org"][0]).Should(Equal(net.ParseIP("192.168.178.3")))
				Expect(config.CustomDNS.Mapping.HostIPs["multiple.ips"][0]).Should(Equal(net.ParseIP("192.168.178.3")))
				Expect(config.CustomDNS.Mapping.HostIPs["multiple.ips"][1]).Should(Equal(net.ParseIP("192.168.178.4")))
				Expect(config.CustomDNS.Mapping.HostIPs["multiple.ips"][2]).Should(Equal(
					net.ParseIP("2001:0db8:85a3:08d3:1319:8a2e:0370:7344")))
				Expect(config.Conditional.Mapping.Upstreams).Should(HaveLen(2))
				Expect(config.Conditional.Mapping.Upstreams["fritz.box"]).Should(HaveLen(1))
				Expect(config.Conditional.Mapping.Upstreams["multiple.resolvers"]).Should(HaveLen(2))
				Expect(config.ClientLookup.Upstream.Host).Should(Equal("192.168.178.1"))
				Expect(config.ClientLookup.SingleNameOrder).Should(Equal([]uint{2, 1}))
				Expect(config.Blocking.BlackLists).Should(HaveLen(2))
				Expect(config.Blocking.WhiteLists).Should(HaveLen(1))
				Expect(config.Blocking.ClientGroupsBlock).Should(HaveLen(2))
				Expect(config.Blocking.BlockTTL).Should(Equal(Duration(time.Minute)))
				Expect(config.Blocking.RefreshPeriod).Should(Equal(Duration(2 * time.Hour)))

				Expect(config.Caching.MaxCachingTime).Should(Equal(Duration(0)))
				Expect(config.Caching.MinCachingTime).Should(Equal(Duration(0)))

				Expect(GetConfig()).Should(Not(BeNil()))

			})
		})
		When("config file is malformed", func() {
			It("should log with fatal and exit", func() {

				dir, err := ioutil.TempDir("", "blocky")
				defer os.Remove(dir)
				Expect(err).Should(Succeed())
				err = os.Chdir(dir)
				Expect(err).Should(Succeed())
				err = ioutil.WriteFile("config.yml", []byte("malformed_config"), 0600)
				Expect(err).Should(Succeed())

				helpertest.ShouldLogFatal(func() {
					LoadConfig("config.yml", true)
				})
			})
		})
		When("duration is in wrong format", func() {
			It("should log with fatal and exit", func() {
				cfg := Config{}
				data :=
					`blocking:
  refreshPeriod: wrongduration`
				helpertest.ShouldLogFatal(func() {
					unmarshalConfig([]byte(data), cfg)
				})
			})
		})
		When("CustomDNS hast wrong IP defined", func() {
			It("should log with fatal and exit", func() {
				cfg := Config{}
				data :=
					`customDNS:
  mapping:
    someDomain: 192.168.178.WRONG`
				helpertest.ShouldLogFatal(func() {
					unmarshalConfig([]byte(data), cfg)
				})
			})
		})
		When("Conditional mapping hast wrong defined upstreams", func() {
			It("should log with fatal and exit", func() {
				cfg := Config{}
				data :=
					`conditional:
  mapping:
    multiple.resolvers: 192.168.178.1,wrongprotocol:4.4.4.4:53`
				helpertest.ShouldLogFatal(func() {
					unmarshalConfig([]byte(data), cfg)
				})
			})
		})
		When("Wrong upstreams are defined", func() {
			It("should log with fatal and exit", func() {
				cfg := Config{}
				data :=
					`upstream:
  default:
    - 8.8.8.8
    - wrongprotocol:8.8.4.4
    - 1.1.1.1`
				helpertest.ShouldLogFatal(func() {
					unmarshalConfig([]byte(data), cfg)
				})
			})
		})

		When("config is not YAML", func() {
			It("should log with fatal and exit", func() {
				cfg := Config{}
				data :=
					`///`
				helpertest.ShouldLogFatal(func() {
					unmarshalConfig([]byte(data), cfg)
				})
			})
		})

		When("TlsPort is defined", func() {
			It("certFile/keyFile must be set", func() {

				By("certFile/keyFile not set", func() {
					c := &Config{
						TLSPorts: ListenConfig{"953"},
					}
					helpertest.ShouldLogFatal(func() {
						validateConfig(c)
					})
				})

				By("certFile/keyFile set", func() {
					c := &Config{
						TLSPorts: ListenConfig{"953"},
						KeyFile:  "key",
						CertFile: "cert",
					}
					validateConfig(c)
				})
			})
		})

		When("HttpsPort is defined", func() {
			It("certFile/keyFile must be set", func() {

				By("certFile/keyFile not set", func() {
					c := &Config{
						HTTPSPorts: ListenConfig{"443"},
					}
					helpertest.ShouldLogFatal(func() {
						validateConfig(c)
					})
				})

				By("certFile/keyFile set", func() {
					c := &Config{
						TLSPorts: ListenConfig{"443"},
						KeyFile:  "key",
						CertFile: "cert",
					}
					validateConfig(c)
				})
			})
		})

		When("config directory does not exist", func() {
			It("should log with fatal and exit if config is mandatory", func() {
				err := os.Chdir("../..")
				Expect(err).Should(Succeed())

				defer func() { Log().ExitFunc = nil }()

				var fatal bool

				Log().ExitFunc = func(int) { fatal = true }
				LoadConfig("config.yml", true)

				Expect(fatal).Should(BeTrue())
			})

			It("should use default config if config is not mandatory", func() {
				err := os.Chdir("../..")
				Expect(err).Should(Succeed())

				LoadConfig("config.yml", false)

				Expect(config.LogLevel).Should(Equal(LevelInfo))
			})
		})
	})

	Describe("YAML parsing", func() {
		Context("upstream", func() {
			It("should create the upstream struct with data", func() {
				u := &Upstream{}
				err := u.UnmarshalYAML(func(i interface{}) error {
					*i.(*string) = "tcp+udp:1.2.3.4"
					return nil

				})
				Expect(err).Should(Succeed())
				Expect(u.Net).Should(Equal(NetProtocolTcpUdp))
				Expect(u.Host).Should(Equal("1.2.3.4"))
				Expect(u.Port).Should(BeNumerically("==", 53))
			})

			It("should fail if the upstream is in wrong format", func() {
				u := &Upstream{}
				err := u.UnmarshalYAML(func(i interface{}) error {
					return errors.New("some err")

				})
				Expect(err).Should(HaveOccurred())
			})
		})
		Context("ListenConfig", func() {
			It("should parse and split valid string config", func() {
				l := &ListenConfig{}
				err := l.UnmarshalYAML(func(i interface{}) error {
					*i.(*string) = "55,:56"
					return nil
				})
				Expect(err).Should(Succeed())
				Expect(*l).Should(HaveLen(2))
				Expect(*l).Should(ContainElements("55", ":56"))
			})
			It("should fail on error", func() {
				l := &ListenConfig{}
				err := l.UnmarshalYAML(func(i interface{}) error {
					return errors.New("some err")
				})
				Expect(err).Should(HaveOccurred())
			})
		})
		Context("Duration", func() {
			It("should parse duration with unit", func() {
				d := Duration(0)
				err := d.UnmarshalYAML(func(i interface{}) error {
					*i.(*string) = "1m20s"
					return nil
				})
				Expect(err).Should(Succeed())
				Expect(d).Should(Equal(Duration(80 * time.Second)))
				Expect(d.String()).Should(Equal("1 minute 20 seconds"))
			})
			It("should fail if duration is in wrong format", func() {
				d := Duration(0)
				err := d.UnmarshalYAML(func(i interface{}) error {
					*i.(*string) = "wrong"
					return nil
				})
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError("time: invalid duration \"wrong\""))

			})
			It("should fail if wrong YAML format", func() {
				d := Duration(0)
				err := d.UnmarshalYAML(func(i interface{}) error {
					return errors.New("some err")
				})
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError("some err"))
			})

		})
		Context("ConditionalUpstreamMapping", func() {
			It("Should parse config as map", func() {
				c := &ConditionalUpstreamMapping{}
				err := c.UnmarshalYAML(func(i interface{}) error {
					*i.(*map[string]string) = map[string]string{"key": "1.2.3.4"}
					return nil
				})
				Expect(err).Should(Succeed())
				Expect(c.Upstreams).Should(HaveLen(1))
				Expect(c.Upstreams["key"]).Should(HaveLen(1))
				Expect(c.Upstreams["key"][0]).Should(Equal(Upstream{
					Net: NetProtocolTcpUdp, Host: "1.2.3.4", Port: 53}))
			})
			It("should fail if wrong YAML format", func() {
				c := &ConditionalUpstreamMapping{}
				err := c.UnmarshalYAML(func(i interface{}) error {
					return errors.New("some err")
				})
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError("some err"))
			})
		})
		Context("CustomDNSMapping", func() {
			It("Should parse config as map", func() {
				c := &CustomDNSMapping{}
				err := c.UnmarshalYAML(func(i interface{}) error {
					*i.(*map[string]string) = map[string]string{"key": "1.2.3.4"}
					return nil
				})
				Expect(err).Should(Succeed())
				Expect(c.HostIPs).Should(HaveLen(1))
				Expect(c.HostIPs["key"]).Should(HaveLen(1))
				Expect(c.HostIPs["key"][0]).Should(Equal(net.ParseIP("1.2.3.4")))
			})
			It("should fail if wrong YAML format", func() {
				c := &CustomDNSMapping{}
				err := c.UnmarshalYAML(func(i interface{}) error {
					return errors.New("some err")
				})
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError("some err"))
			})
		})
	})

	DescribeTable("parse upstream string",
		func(in string, wantResult Upstream, wantErr bool) {
			result, err := ParseUpstream(in)
			if wantErr {
				Expect(err).Should(HaveOccurred(), in)
			} else {
				Expect(err).Should(Succeed(), in)
			}
			Expect(result).Should(Equal(wantResult), in)
		},
		Entry("udp+tcp with port",
			"4.4.4.4:531",
			Upstream{Net: NetProtocolTcpUdp, Host: "4.4.4.4", Port: 531},
			false),
		Entry("udp+tcü without port, use default",
			"4.4.4.4",
			Upstream{Net: NetProtocolTcpUdp, Host: "4.4.4.4", Port: 53},
			false),
		Entry("udp+tcp with port",
			"tcp+udp:4.4.4.4:4711",
			Upstream{Net: NetProtocolTcpUdp, Host: "4.4.4.4", Port: 4711},
			false),
		Entry("tcp without port, use default",
			"4.4.4.4",
			Upstream{Net: NetProtocolTcpUdp, Host: "4.4.4.4", Port: 53},
			false),
		Entry("tcp-tls without port, use default",
			"tcp-tls:4.4.4.4",
			Upstream{Net: NetProtocolTcpTls, Host: "4.4.4.4", Port: 853},
			false),
		Entry("DoH without port, use default",
			"https:4.4.4.4",
			Upstream{Net: NetProtocolHttps, Host: "4.4.4.4", Port: 443},
			false),
		Entry("DoH with port",
			"https:4.4.4.4:888",
			Upstream{Net: NetProtocolHttps, Host: "4.4.4.4", Port: 888},
			false),
		Entry("DoH named",
			"https://dns.google/dns-query",
			Upstream{Net: NetProtocolHttps, Host: "dns.google", Port: 443, Path: "/dns-query"},
			false),
		Entry("DoH named, path with multiple slashes",
			"https://dns.google/dns-query/a/b",
			Upstream{Net: NetProtocolHttps, Host: "dns.google", Port: 443, Path: "/dns-query/a/b"},
			false),
		Entry("DoH named with port",
			"https://dns.google:888/dns-query",
			Upstream{Net: NetProtocolHttps, Host: "dns.google", Port: 888, Path: "/dns-query"},
			false),
		Entry("empty",
			"",
			Upstream{Net: 0},
			true),
		Entry("udpIpv6WithPort",
			"tcp+udp:[fd00::6cd4:d7e0:d99d:2952]:53",
			Upstream{Net: NetProtocolTcpUdp, Host: "fd00::6cd4:d7e0:d99d:2952", Port: 53},
			false),
		Entry("udpIpv6WithPort2",
			"[2001:4860:4860::8888]:53",
			Upstream{Net: NetProtocolTcpUdp, Host: "2001:4860:4860::8888", Port: 53},
			false),
		Entry("default net, default port",
			"1.1.1.1",
			Upstream{Net: NetProtocolTcpUdp, Host: "1.1.1.1", Port: 53},
			false),
		Entry("wrong host name",
			"host$name",
			Upstream{},
			true),
		Entry("default net with port",
			"1.1.1.1:153",
			Upstream{Net: NetProtocolTcpUdp, Host: "1.1.1.1", Port: 153},
			false),
		Entry("with negative port",
			"tcp:4.4.4.4:-1",
			nil,
			true),
		Entry("with invalid port",
			"tcp:4.4.4.4:65536",
			nil,
			true),
		Entry("with not numeric port",
			"tcp:4.4.4.4:A636",
			nil,
			true),
		Entry("with wrong protocol",
			"bla:4.4.4.4:53",
			nil,
			true),
		Entry("tcp+udp",
			"tcp+udp:1.1.1.1:53",
			Upstream{Net: NetProtocolTcpUdp, Host: "1.1.1.1", Port: 53},
			false),
		Entry("tcp+udp default port",
			"tcp+udp:1.1.1.1",
			Upstream{Net: NetProtocolTcpUdp, Host: "1.1.1.1", Port: 53},
			false),
		Entry("defaultIpv6Short",
			"2620:fe::fe",
			Upstream{Net: NetProtocolTcpUdp, Host: "2620:fe::fe", Port: 53},
			false),
		Entry("defaultIpv6Short2",
			"2620:fe::9",
			Upstream{Net: NetProtocolTcpUdp, Host: "2620:fe::9", Port: 53},
			false),
		Entry("defaultIpv6WithPort",
			"[2620:fe::9]:55",
			Upstream{Net: NetProtocolTcpUdp, Host: "2620:fe::9", Port: 55},
			false),
	)
})
