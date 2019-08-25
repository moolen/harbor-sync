package controllers

import (
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Adapter", func() {
	It("should send events", func(done Done) {
		var cl client.Client
		var ad Adapter
		input1 := make(chan struct{})
		input := []<-chan struct{}{input1}
		scheme := runtime.NewScheme()
		fooCFG := crdv1.HarborSyncConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "1", Namespace: "1"},
		}
		barCFG := crdv1.HarborSyncConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "2", Namespace: "2"},
		}
		bazCFG := crdv1.HarborSyncConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "3", Namespace: "3"},
		}
		log := zap.Logger(true)
		crdv1.AddToScheme(scheme)
		cl = fake.NewFakeClientWithScheme(scheme, &fooCFG, &barCFG, &bazCFG)

		ad = NewAdapter(cl, log, input)
		out := ad.Run()

		input1 <- struct{}{}

		evt1 := <-out
		evt2 := <-out
		evt3 := <-out

		Expect(evt1.Meta.(*metav1.ObjectMeta).Name).To(Equal("1"))
		Expect(evt2.Meta.(*metav1.ObjectMeta).Name).To(Equal("2"))
		Expect(evt3.Meta.(*metav1.ObjectMeta).Name).To(Equal("3"))
		close(done)
	})
})
