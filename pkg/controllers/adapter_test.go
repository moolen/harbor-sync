/*
Copyright 2019 The Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moolen/harbor-sync/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Adapter", func() {

	BeforeEach(func() {
		test.EnsureHarborSyncConfig(k8sClient, "bar")
		test.EnsureHarborSyncConfig(k8sClient, "baz")
		test.EnsureHarborSyncConfig(k8sClient, "foo")
	})

	AfterEach(func() {
		test.DeleteHarborSyncConfig(k8sClient, "bar")
		test.DeleteHarborSyncConfig(k8sClient, "baz")
		test.DeleteHarborSyncConfig(k8sClient, "foo")
	})

	It("should send events", func(done Done) {
		var ad Adapter
		input1 := make(chan struct{})
		input := []<-chan struct{}{input1}

		log := zap.Logger(false)

		ad = NewAdapter(k8sClient, log, input)
		out := ad.Run()

		input1 <- struct{}{}

		evt1 := <-out
		evt2 := <-out
		evt3 := <-out

		Expect(evt1.Meta.(*metav1.ObjectMeta).Name).To(Equal("bar"))
		Expect(evt2.Meta.(*metav1.ObjectMeta).Name).To(Equal("baz"))
		Expect(evt3.Meta.(*metav1.ObjectMeta).Name).To(Equal("foo"))
		close(done)
	})
})
