package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "gce.go",
        "gce_addresses.go",
        "gce_annotations.go",
        "gce_backendservice.go",
        "gce_cert.go",
        "gce_clusterid.go",
        "gce_clusters.go",
        "gce_disks.go",
        "gce_firewall.go",
        "gce_forwardingrule.go",
        "gce_healthchecks.go",
        "gce_instancegroup.go",
        "gce_instances.go",
        "gce_loadbalancer.go",
        "gce_loadbalancer_external.go",
        "gce_loadbalancer_internal.go",
        "gce_loadbalancer_naming.go",
        "gce_op.go",
        "gce_routes.go",
        "gce_targetpool.go",
        "gce_targetproxy.go",
        "gce_urlmap.go",
        "gce_util.go",
        "gce_zones.go",
        "metrics.go",
        "token_source.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//pkg/api/v1/service:go_default_library",
        "//pkg/client/clientset_generated/clientset:go_default_library",
        "//pkg/cloudprovider:go_default_library",
        "//pkg/controller:go_default_library",
        "//pkg/kubelet/apis:go_default_library",
        "//pkg/master/ports:go_default_library",
        "//pkg/util/net/sets:go_default_library",
        "//pkg/util/version:go_default_library",
        "//pkg/volume:go_default_library",
        "//vendor/cloud.google.com/go/compute/metadata:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/golang.org/x/oauth2/google:go_default_library",
        "//vendor/google.golang.org/api/compute/v0.beta:go_default_library",
        "//vendor/google.golang.org/api/compute/v1:go_default_library",
        "//vendor/google.golang.org/api/container/v1:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/gopkg.in/gcfg.v1:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/fields:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/types:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/wait:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/client-go/util/flowcontrol:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "gce_disks_test.go",
        "gce_healthchecks_test.go",
        "gce_test.go",
    ],
    library = ":go_default_library",
    tags = ["automanaged"],
    deps = [
        "//pkg/cloudprovider:go_default_library",
        "//pkg/kubelet/apis:go_default_library",
        "//vendor/google.golang.org/api/compute/v1:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
