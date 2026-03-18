package scoring

import "strings"

// NormalizeRole maps various role titles to standard roles
func NormalizeRole(title string) string {
	lower := strings.ToLower(strings.TrimSpace(title))
	for pattern, normalized := range roleNormalization {
		if strings.Contains(lower, pattern) {
			return normalized
		}
	}
	return lower
}

// RoleFamily returns the role family for a given normalized role
func RoleFamily(role string) string {
	for family, roles := range roleFamilies {
		for _, r := range roles {
			if strings.Contains(role, r) {
				return family
			}
		}
	}
	return "other"
}

// NormalizeSkill maps skill synonyms to standard names
func NormalizeSkill(skill string) string {
	lower := strings.ToLower(strings.TrimSpace(skill))
	if normalized, ok := skillNormalization[lower]; ok {
		return normalized
	}
	return lower
}

// NormalizeCompany maps company name variants to standard names
func NormalizeCompany(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	for pattern, normalized := range companyNormalization {
		if strings.Contains(lower, pattern) {
			return normalized
		}
	}
	return lower
}

// DetermineSeniority based on total experience months and evidence
func DetermineSeniority(totalMonths int, hasLeadership bool) string {
	switch {
	case totalMonths < 12:
		return "intern"
	case totalMonths < 30:
		return "junior"
	case totalMonths < 60:
		return "middle"
	case totalMonths < 96:
		if hasLeadership {
			return "lead"
		}
		return "senior"
	default:
		if hasLeadership {
			return "lead"
		}
		return "senior"
	}
}

var roleNormalization = map[string]string{
	"golang":                  "backend developer",
	"backend engineer":        "backend developer",
	"backend разработчик":     "backend developer",
	"бэкенд разработчик":     "backend developer",
	"golang developer":        "backend developer",
	"go developer":            "backend developer",
	"java developer":          "backend developer",
	"python developer":        "backend developer",
	"node.js developer":       "backend developer",
	"php developer":           "backend developer",
	"c# developer":            "backend developer",
	".net developer":          "backend developer",
	"software engineer":       "software engineer",
	"full-stack":              "fullstack developer",
	"full stack":              "fullstack developer",
	"fullstack":               "fullstack developer",
	"frontend engineer":       "frontend developer",
	"frontend разработчик":    "frontend developer",
	"фронтенд разработчик":   "frontend developer",
	"react developer":         "frontend developer",
	"vue developer":           "frontend developer",
	"angular developer":       "frontend developer",
	"ui developer":            "frontend developer",
	"mobile developer":        "mobile developer",
	"ios developer":           "mobile developer",
	"android developer":       "mobile developer",
	"flutter developer":       "mobile developer",
	"react native developer":  "mobile developer",
	"data scientist":          "data scientist",
	"data analyst":            "data analyst",
	"data engineer":           "data engineer",
	"ml engineer":             "ml engineer",
	"machine learning":        "ml engineer",
	"qa engineer":             "qa engineer",
	"test engineer":           "qa engineer",
	"quality assurance":       "qa engineer",
	"тестировщик":             "qa engineer",
	"devops engineer":         "devops engineer",
	"sre":                     "devops engineer",
	"site reliability":        "devops engineer",
	"infrastructure":          "devops engineer",
	"product manager":         "product manager",
	"project manager":         "project manager",
	"ui/ux":                   "designer",
	"ux designer":             "designer",
	"ui designer":             "designer",
	"graphic designer":        "designer",
	"дизайнер":                "designer",
	"system administrator":    "system administrator",
	"системный администратор": "system administrator",
	"dba":                     "database administrator",
	"database administrator":  "database administrator",
	"technical writer":        "technical writer",
	"business analyst":        "business analyst",
	"бизнес-аналитик":         "business analyst",
	"cyber security":          "security engineer",
	"security engineer":       "security engineer",
	"информационная безопас":  "security engineer",
	"game developer":          "game developer",
	"embedded":                "embedded developer",
	"tech lead":               "tech lead",
	"team lead":               "tech lead",
	"тимлид":                  "tech lead",
	"тех лид":                 "tech lead",
	"архитектор":              "software architect",
	"architect":               "software architect",
	"cto":                     "cto",
	"vp of engineering":       "vp engineering",
	"head of":                 "head of engineering",
	"marketing":               "marketing specialist",
	"маркетолог":              "marketing specialist",
	"hr manager":              "hr manager",
	"recruiter":               "recruiter",
	"рекрутер":                "recruiter",
}

var roleFamilies = map[string][]string{
	"backend":  {"backend", "golang", "java developer", "python developer", "node.js", "php", "c#", ".net", "software engineer", "fullstack"},
	"frontend": {"frontend", "react developer", "vue", "angular", "ui developer", "fullstack"},
	"mobile":   {"mobile", "ios", "android", "flutter", "react native"},
	"data":     {"data scientist", "data analyst", "data engineer", "ml engineer", "machine learning"},
	"qa":       {"qa", "test", "quality assurance"},
	"devops":   {"devops", "sre", "site reliability", "infrastructure", "system administrator"},
	"pm":       {"product manager", "project manager"},
	"design":   {"designer", "ui/ux", "ux", "graphic"},
	"security": {"security", "cyber"},
	"other":    {"technical writer", "business analyst", "marketing", "hr", "recruiter", "game developer", "embedded"},
}

var skillNormalization = map[string]string{
	"postgres":      "postgresql",
	"postgresql":    "postgresql",
	"mongo":         "mongodb",
	"mongodb":       "mongodb",
	"k8s":           "kubernetes",
	"kubernetes":    "kubernetes",
	"ci/cd":         "ci_cd",
	"cicd":          "ci_cd",
	"ci cd":         "ci_cd",
	"grpc":          "grpc",
	"rest api":      "rest_api",
	"restful":       "rest_api",
	"rest":          "rest_api",
	"docker":        "docker",
	"react":         "react",
	"reactjs":       "react",
	"react.js":      "react",
	"vue":           "vuejs",
	"vue.js":        "vuejs",
	"vuejs":         "vuejs",
	"angular":       "angular",
	"typescript":    "typescript",
	"ts":            "typescript",
	"javascript":    "javascript",
	"js":            "javascript",
	"python":        "python",
	"golang":        "go",
	"go":            "go",
	"java":          "java",
	"c#":            "csharp",
	"c sharp":       "csharp",
	".net":          "dotnet",
	"dotnet":        "dotnet",
	"asp.net":       "dotnet",
	"node":          "nodejs",
	"node.js":       "nodejs",
	"nodejs":        "nodejs",
	"express":       "expressjs",
	"express.js":    "expressjs",
	"flutter":       "flutter",
	"dart":          "dart",
	"kotlin":        "kotlin",
	"swift":         "swift",
	"redis":         "redis",
	"kafka":         "kafka",
	"rabbitmq":      "rabbitmq",
	"aws":           "aws",
	"gcp":           "gcp",
	"azure":         "azure",
	"terraform":     "terraform",
	"nginx":         "nginx",
	"linux":         "linux",
	"git":           "git",
	"mysql":         "mysql",
	"sqlite":        "sqlite",
	"elasticsearch": "elasticsearch",
	"graphql":       "graphql",
	"next.js":       "nextjs",
	"nextjs":        "nextjs",
	"tailwind":      "tailwind",
	"tailwind css":  "tailwind",
	"figma":         "figma",
	"storybook":     "storybook",
	"firebase":      "firebase",
	"bloc":          "bloc",
	"riverpod":      "riverpod",
	"pandas":        "pandas",
	"numpy":         "numpy",
	"scikit-learn":  "scikit_learn",
	"sklearn":       "scikit_learn",
	"tableau":       "tableau",
	"power bi":      "powerbi",
	"powerbi":       "powerbi",
	"jupyter":       "jupyter",
}

var companyNormalization = map[string]string{
	"yandex":      "yandex",
	"яндекс":      "yandex",
	"google":      "google",
	"meta":        "meta",
	"facebook":    "meta",
	"apple":       "apple",
	"amazon":      "amazon",
	"microsoft":   "microsoft",
	"epam":        "epam",
	"kaspersky":   "kaspersky",
	"jetbrains":   "jetbrains",
	"tinkoff":     "tinkoff",
	"t-bank":      "tinkoff",
	"vk":          "vk",
	"revolut":     "revolut",
	"stripe":      "stripe",
	"payme":       "payme",
	"click":       "click",
	"uzum":        "uzum",
	"mytaxi":      "mytaxi",
	"yandex go":   "mytaxi",
	"billz":       "billz",
	"humans":      "humans",
	"apelsin":     "apelsin",
	"osontaxi":    "osontaxi",
	"korzinka":    "korzinka",
	"artel":       "artel",
	"uzauto":      "uzauto",
	"ucell":       "ucell",
	"beeline":     "beeline_uz",
	"it park":     "it_park",
	"mediapark":   "mediapark",
	"kapitalbank": "kapitalbank",
	"fido biznes": "fido_biznes",
	"fido":        "fido_biznes",
}

// Domain normalization
var domainKeywords = map[string][]string{
	"ecommerce":  {"ecommerce", "e-commerce", "online store", "marketplace", "магазин", "маркетплейс", "shop", "retail"},
	"payments":   {"payment", "billing", "checkout", "acquiring", "платёж", "оплата", "transaction"},
	"fintech":    {"fintech", "banking", "bank", "lending", "финтех", "банк", "кредит", "financial"},
	"logistics":  {"logistics", "delivery", "shipping", "логистика", "доставка", "transportation"},
	"healthcare": {"health", "medical", "medicine", "здоровье", "медицин"},
	"edtech":     {"edtech", "education", "learning", "course", "образовани", "обучени"},
	"enterprise": {"crm", "erp", "enterprise", "saas", "b2b"},
	"analytics":  {"analytics", "bi", "data warehouse", "dwh", "аналитик"},
	"gaming":     {"gaming", "game", "gamedev", "игр"},
	"telecom":    {"telecom", "телеком", "mobile operator", "связь"},
	"media":      {"media", "news", "content", "медиа", "контент", "новост"},
	"taxi":       {"taxi", "ride", "такси"},
}

// ExtractDomains finds matching domains from text
func ExtractDomains(texts ...string) []string {
	combined := strings.ToLower(strings.Join(texts, " "))
	var domains []string
	seen := make(map[string]bool)
	for domain, keywords := range domainKeywords {
		for _, kw := range keywords {
			if strings.Contains(combined, kw) && !seen[domain] {
				domains = append(domains, domain)
				seen[domain] = true
				break
			}
		}
	}
	return domains
}
