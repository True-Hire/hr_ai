package casbin

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	pgxadapter "github.com/pckhoi/casbin-pgx-adapter/v3"
)

const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

func NewEnforcer(databaseURL string) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("create casbin model: %w", err)
	}

	a, err := pgxadapter.NewAdapter(databaseURL, pgxadapter.WithDatabase("hr_ai_db"), pgxadapter.WithTableName("casbin_rule"))
	if err != nil {
		return nil, fmt.Errorf("create casbin adapter: %w", err)
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}

	// Seed default policies if none exist
	existingPolicies, _ := e.GetPolicy()
	if len(existingPolicies) == 0 {
		policies := [][]string{
			{"hr", "vacancies", "create"},
			{"hr", "vacancies", "read"},
			{"hr", "vacancies", "update"},
			{"hr", "vacancies", "delete"},
			{"user", "vacancies", "read"},
			{"hr", "companies", "create"},
			{"hr", "companies", "read"},
			{"hr", "companies", "update"},
			{"hr", "companies", "delete"},
			{"user", "companies", "read"},
		}
		for _, p := range policies {
			if _, err := e.AddPolicy(p); err != nil {
				return nil, fmt.Errorf("add casbin policy: %w", err)
			}
		}
		if err := e.SavePolicy(); err != nil {
			return nil, fmt.Errorf("save casbin policies: %w", err)
		}
	}

	return e, nil
}
