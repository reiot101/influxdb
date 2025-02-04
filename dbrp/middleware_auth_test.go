package dbrp_test

import (
	"context"
	"github.com/influxdata/influxdb/v2/kit/platform"
	"github.com/influxdata/influxdb/v2/kit/platform/errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb/v2"
	influxdbcontext "github.com/influxdata/influxdb/v2/context"
	"github.com/influxdata/influxdb/v2/dbrp"
	"github.com/influxdata/influxdb/v2/mock"
	influxdbtesting "github.com/influxdata/influxdb/v2/testing"
)

func TestAuth_FindByID(t *testing.T) {
	type fields struct {
		service influxdb.DBRPMappingService
	}
	type args struct {
		orgID      platform.ID
		id         platform.ID
		permission influxdb.Permission
	}
	type wants struct {
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "authorized to access id by org id",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(1),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "authorized to access id by id",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(1),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type: influxdb.BucketsResourceType,
						ID:   influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "unauthorized to access id by org id",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(2),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 2,
			},
			wants: wants{
				err: &errors.Error{
					Msg:  "read:orgs/0000000000000002/buckets/0000000000000001 is unauthorized",
					Code: errors.EUnauthorized,
				},
			},
		},
		{
			name: "unauthorized to access id by id",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(1),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type: influxdb.DBRPResourceType,
						ID:   influxdbtesting.IDPtr(2),
					},
				},
				id:    1,
				orgID: 2,
			},
			wants: wants{
				err: &errors.Error{
					Msg:  "read:orgs/0000000000000002/buckets/0000000000000001 is unauthorized",
					Code: errors.EUnauthorized,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := dbrp.NewAuthorizedService(tt.fields.service)

			ctx := context.Background()
			ctx = influxdbcontext.SetAuthorizer(ctx, mock.NewMockAuthorizer(false, []influxdb.Permission{tt.args.permission}))

			_, err := s.FindByID(ctx, tt.args.orgID, tt.args.id)
			influxdbtesting.ErrorsEqual(t, err, tt.wants.err)
		})
	}
}

func TestAuth_FindMany(t *testing.T) {
	type fields struct {
		service influxdb.DBRPMappingService
	}
	type args struct {
		filter      influxdb.DBRPMappingFilter
		permissions []influxdb.Permission
	}
	type wants struct {
		err error
		ms  []*influxdb.DBRPMapping
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "no result",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindManyFn: func(ctx context.Context, dbrp influxdb.DBRPMappingFilter, opts ...influxdb.FindOptions) ([]*influxdb.DBRPMapping, int, error) {
						return []*influxdb.DBRPMapping{
							{
								ID:             1,
								OrganizationID: 1,
								BucketID:       1,
							},
							{
								ID:             2,
								OrganizationID: 1,
								BucketID:       2,
							},
							{
								ID:             3,
								OrganizationID: 2,
								BucketID:       3,
							},
							{
								ID:             4,
								OrganizationID: 3,
								BucketID:       4,
							},
						}, 4, nil
					},
				},
			},
			args: args{
				permissions: []influxdb.Permission{{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.DBRPResourceType,
						OrgID: influxdbtesting.IDPtr(42),
					},
				}},
				filter: influxdb.DBRPMappingFilter{},
			},
			wants: wants{
				err: nil,
				ms:  []*influxdb.DBRPMapping{},
			},
		},
		{
			name: "partial",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindManyFn: func(ctx context.Context, dbrp influxdb.DBRPMappingFilter, opts ...influxdb.FindOptions) ([]*influxdb.DBRPMapping, int, error) {
						return []*influxdb.DBRPMapping{
							{
								ID:             1,
								OrganizationID: 1,
								BucketID:       1,
							},
							{
								ID:             2,
								OrganizationID: 1,
								BucketID:       2,
							},
							{
								ID:             3,
								OrganizationID: 2,
								BucketID:       3,
							},
							{
								ID:             4,
								OrganizationID: 3,
								BucketID:       4,
							},
						}, 4, nil
					},
				},
			},
			args: args{
				permissions: []influxdb.Permission{{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				}},
				filter: influxdb.DBRPMappingFilter{},
			},
			wants: wants{
				err: nil,
				ms: []*influxdb.DBRPMapping{
					{
						ID:             1,
						OrganizationID: 1,
						BucketID:       1,
					},
					{
						ID:             2,
						OrganizationID: 1,
						BucketID:       2,
					},
				},
			},
		},
		{
			name: "all",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindManyFn: func(ctx context.Context, dbrp influxdb.DBRPMappingFilter, opts ...influxdb.FindOptions) ([]*influxdb.DBRPMapping, int, error) {
						return []*influxdb.DBRPMapping{
							{
								ID:             1,
								OrganizationID: 1,
								BucketID:       1,
							},
							{
								ID:             2,
								OrganizationID: 1,
								BucketID:       2,
							},
							{
								ID:             3,
								OrganizationID: 2,
								BucketID:       3,
							},
							{
								ID:             4,
								OrganizationID: 3,
								BucketID:       4,
							},
						}, 4, nil
					},
				},
			},
			args: args{
				permissions: []influxdb.Permission{
					{
						Action: "read",
						Resource: influxdb.Resource{
							Type:  influxdb.BucketsResourceType,
							OrgID: influxdbtesting.IDPtr(1),
						},
					},
					{
						Action: "read",
						Resource: influxdb.Resource{
							Type:  influxdb.BucketsResourceType,
							OrgID: influxdbtesting.IDPtr(2),
						},
					},
					{
						Action: "read",
						Resource: influxdb.Resource{
							Type:  influxdb.BucketsResourceType,
							OrgID: influxdbtesting.IDPtr(3),
						},
					},
				},
				filter: influxdb.DBRPMappingFilter{},
			},
			wants: wants{
				err: nil,
				ms: []*influxdb.DBRPMapping{
					{
						ID:             1,
						OrganizationID: 1,
						BucketID:       1,
					},
					{
						ID:             2,
						OrganizationID: 1,
						BucketID:       2,
					},
					{
						ID:             3,
						OrganizationID: 2,
						BucketID:       3,
					},
					{
						ID:             4,
						OrganizationID: 3,
						BucketID:       4,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := dbrp.NewAuthorizedService(tt.fields.service)

			ctx := context.Background()
			ctx = influxdbcontext.SetAuthorizer(ctx, mock.NewMockAuthorizer(false, tt.args.permissions))

			gots, ngots, err := s.FindMany(ctx, tt.args.filter)
			if ngots != len(gots) {
				t.Errorf("got wrong number back")
			}
			influxdbtesting.ErrorsEqual(t, err, tt.wants.err)
			if diff := cmp.Diff(tt.wants.ms, gots, influxdbtesting.DBRPMappingCmpOptions...); diff != "" {
				t.Errorf("unexpected result -want/+got:\n\t%s", diff)
			}
		})
	}
}

func TestAuth_Create(t *testing.T) {
	type fields struct {
		service influxdb.DBRPMappingService
	}
	type args struct {
		m          influxdb.DBRPMapping
		permission influxdb.Permission
	}
	type wants struct {
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "authorized",
			fields: fields{
				service: &mock.DBRPMappingService{},
			},
			args: args{
				m: influxdb.DBRPMapping{
					ID:             1,
					OrganizationID: 1,
					BucketID:       2,
				},
				permission: influxdb.Permission{
					Action: "write",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
						ID:    influxdbtesting.IDPtr(2),
					},
				},
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "unauthorized",
			fields: fields{
				service: &mock.DBRPMappingService{},
			},
			args: args{
				m: influxdb.DBRPMapping{
					ID:             1,
					OrganizationID: 1,
					BucketID:       2,
				},
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
						ID:    influxdbtesting.IDPtr(2),
					},
				},
			},
			wants: wants{
				err: &errors.Error{
					Msg:  "write:orgs/0000000000000001/buckets/0000000000000002 is unauthorized",
					Code: errors.EUnauthorized,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := dbrp.NewAuthorizedService(tt.fields.service)

			ctx := context.Background()
			ctx = influxdbcontext.SetAuthorizer(ctx, mock.NewMockAuthorizer(false, []influxdb.Permission{tt.args.permission}))

			err := s.Create(ctx, &tt.args.m)
			influxdbtesting.ErrorsEqual(t, err, tt.wants.err)
		})
	}
}

func TestAuth_Update(t *testing.T) {
	type fields struct {
		service influxdb.DBRPMappingService
	}
	type args struct {
		orgID      platform.ID
		id         platform.ID
		permission influxdb.Permission
	}
	type wants struct {
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "authorized",
			fields: fields{
				service: &mock.DBRPMappingService{},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "write",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "unauthorized",
			fields: fields{
				service: &mock.DBRPMappingService{},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: &errors.Error{
					Msg:  "write:orgs/0000000000000001/buckets/0000000000000001 is unauthorized",
					Code: errors.EUnauthorized,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := dbrp.NewAuthorizedService(tt.fields.service)

			ctx := context.Background()
			ctx = influxdbcontext.SetAuthorizer(ctx, mock.NewMockAuthorizer(false, []influxdb.Permission{tt.args.permission}))

			// Does not matter how we update, we only need to check auth.
			err := s.Update(ctx, &influxdb.DBRPMapping{ID: tt.args.id, OrganizationID: tt.args.orgID, BucketID: 1})
			influxdbtesting.ErrorsEqual(t, err, tt.wants.err)
		})
	}
}

func TestAuth_Delete(t *testing.T) {
	type fields struct {
		service influxdb.DBRPMappingService
	}
	type args struct {
		orgID      platform.ID
		id         platform.ID
		permission influxdb.Permission
	}
	type wants struct {
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "authorized",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(1),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "write",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "unauthorized",
			fields: fields{
				service: &mock.DBRPMappingService{
					FindByIDFn: func(_ context.Context, _, _ platform.ID) (*influxdb.DBRPMapping, error) {
						return &influxdb.DBRPMapping{
							OrganizationID: platform.ID(1),
							BucketID:       platform.ID(1),
						}, nil
					},
				},
			},
			args: args{
				permission: influxdb.Permission{
					Action: "read",
					Resource: influxdb.Resource{
						Type:  influxdb.BucketsResourceType,
						OrgID: influxdbtesting.IDPtr(1),
					},
				},
				id:    1,
				orgID: 1,
			},
			wants: wants{
				err: &errors.Error{
					Msg:  "write:orgs/0000000000000001/buckets/0000000000000001 is unauthorized",
					Code: errors.EUnauthorized,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := dbrp.NewAuthorizedService(tt.fields.service)

			ctx := context.Background()
			ctx = influxdbcontext.SetAuthorizer(ctx, mock.NewMockAuthorizer(false, []influxdb.Permission{tt.args.permission}))

			err := s.Delete(ctx, tt.args.orgID, tt.args.id)
			influxdbtesting.ErrorsEqual(t, err, tt.wants.err)
		})
	}
}
