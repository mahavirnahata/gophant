package queue

import "testing"

type testJob struct {
	Name string `json:"name"`
}

func (j *testJob) Handle() error { return nil }

func TestRegistrySerializeDeserialize(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterType(&testJob{}, func() JobHandler { return &testJob{} })

	payload, err := reg.Serialize(&testJob{Name: "x"})
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	job, err := reg.Deserialize(payload)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}

	j := job.(*testJob)
	if j.Name != "x" {
		t.Fatalf("expected name")
	}
}
