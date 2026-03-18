package scoring

import (
	"math"
	"strings"
)

// Clamp value to [0, 1]
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// containsAny checks if text contains any of the keywords
func containsAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// countMatches counts how many keywords appear in text
func countMatches(text string, keywords []string) int {
	lower := strings.ToLower(text)
	count := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			count++
		}
	}
	return count
}

// CalcBackendScore measures how strongly candidate is backend-oriented
func CalcBackendScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)

	// Title match +0.35
	backendTitles := []string{"backend", "server", "бэкенд", "бекенд", "api developer"}
	if containsAny(titleLower, backendTitles) {
		score += 0.35
	}

	// Backend language evidence +0.20
	backendLangs := []string{"go", "java", "python", "csharp", "dotnet", "nodejs", "php", "ruby", "rust"}
	matched := 0
	for _, sk := range skills {
		for _, lang := range backendLangs {
			if sk == lang {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.07, 0.20)
	}

	// DB/broker/API evidence +0.20
	infraSkills := []string{"postgresql", "mysql", "mongodb", "redis", "kafka", "rabbitmq", "grpc", "rest_api", "graphql", "elasticsearch"}
	infraMatched := 0
	for _, sk := range skills {
		for _, infra := range infraSkills {
			if sk == infra {
				infraMatched++
				break
			}
		}
	}
	if infraMatched > 0 {
		score += math.Min(float64(infraMatched)*0.05, 0.20)
	}

	// Backend project evidence +0.25
	backendKeywords := []string{"api", "microservice", "service", "server", "endpoint", "backend", "database", "migration", "queue", "grpc", "rest"}
	for _, text := range experienceTexts {
		matches := countMatches(text, backendKeywords)
		if matches > 0 {
			score += math.Min(float64(matches)*0.05, 0.25)
			break
		}
	}

	return clamp01(score)
}

// CalcFrontendScore measures frontend orientation
func CalcFrontendScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)

	if containsAny(titleLower, []string{"frontend", "front-end", "фронтенд", "ui developer", "web developer"}) {
		score += 0.35
	}

	frontendSkills := []string{"react", "vuejs", "angular", "typescript", "javascript", "nextjs", "tailwind", "html", "css", "storybook", "figma"}
	matched := 0
	for _, sk := range skills {
		for _, fs := range frontendSkills {
			if sk == fs {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.06, 0.25)
	}

	frontendKw := []string{"component", "spa", "responsive", "ui", "ux", "layout", "dom", "css", "frontend", "web app"}
	for _, text := range experienceTexts {
		if countMatches(text, frontendKw) > 0 {
			score += 0.20
			break
		}
	}

	return clamp01(score)
}

// CalcMobileScore measures mobile development orientation
func CalcMobileScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)

	if containsAny(titleLower, []string{"mobile", "ios", "android", "flutter", "react native"}) {
		score += 0.35
	}

	mobileSkills := []string{"flutter", "dart", "kotlin", "swift", "react_native", "ios", "android", "firebase", "bloc", "riverpod"}
	matched := 0
	for _, sk := range skills {
		for _, ms := range mobileSkills {
			if sk == ms {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.08, 0.30)
	}

	mobileKw := []string{"mobile app", "app store", "google play", "push notification", "offline", "мобильн"}
	for _, text := range experienceTexts {
		if countMatches(text, mobileKw) > 0 {
			score += 0.20
			break
		}
	}

	return clamp01(score)
}

// CalcDataScore measures data science/engineering orientation
func CalcDataScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)

	if containsAny(titleLower, []string{"data", "ml", "machine learning", "аналитик"}) {
		score += 0.35
	}

	dataSkills := []string{"pandas", "numpy", "scikit_learn", "tensorflow", "pytorch", "jupyter", "tableau", "powerbi", "spark", "airflow", "hadoop"}
	matched := 0
	for _, sk := range skills {
		for _, ds := range dataSkills {
			if sk == ds {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.08, 0.30)
	}

	dataKw := []string{"model", "prediction", "analytics", "dashboard", "etl", "pipeline", "feature engineering", "a/b test"}
	for _, text := range experienceTexts {
		if countMatches(text, dataKw) > 0 {
			score += 0.20
			break
		}
	}

	return clamp01(score)
}

// CalcQAScore measures QA orientation
func CalcQAScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	if containsAny(strings.ToLower(title), []string{"qa", "test", "quality", "тестировщик"}) {
		score += 0.40
	}
	qaSkills := []string{"selenium", "cypress", "jest", "pytest", "junit", "testng", "postman", "jmeter"}
	for _, sk := range skills {
		for _, qs := range qaSkills {
			if sk == qs {
				score += 0.10
			}
		}
	}
	qaKw := []string{"test case", "bug report", "regression", "automation test", "manual test", "load test"}
	for _, text := range experienceTexts {
		if countMatches(text, qaKw) > 0 {
			score += 0.20
			break
		}
	}
	return clamp01(score)
}

// CalcPMScore measures project/product management orientation
func CalcPMScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	if containsAny(strings.ToLower(title), []string{"product manager", "project manager", "scrum master", "agile coach"}) {
		score += 0.40
	}
	pmKw := []string{"roadmap", "sprint", "backlog", "stakeholder", "user story", "agile", "scrum", "kanban", "delivery", "prioritiz"}
	for _, text := range experienceTexts {
		matches := countMatches(text, pmKw)
		if matches > 0 {
			score += math.Min(float64(matches)*0.08, 0.40)
		}
	}
	return clamp01(score)
}

// CalcDevOpsRoleScore measures how strongly candidate is DevOps-oriented as a role
func CalcDevOpsRoleScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	if containsAny(strings.ToLower(title), []string{"devops", "sre", "infrastructure", "platform engineer", "site reliability"}) {
		score += 0.40
	}
	devopsSkills := []string{"kubernetes", "terraform", "ansible", "aws", "gcp", "azure", "docker", "ci_cd", "prometheus", "grafana"}
	matched := 0
	for _, sk := range skills {
		for _, ds := range devopsSkills {
			if sk == ds {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.06, 0.30)
	}
	devopsKw := []string{"infrastructure", "pipeline", "deployment", "monitoring", "scaling", "terraform", "ansible", "cloud"}
	for _, text := range experienceTexts {
		if countMatches(text, devopsKw) > 0 {
			score += 0.20
			break
		}
	}
	return clamp01(score)
}

// CalcDesignScore measures design orientation
func CalcDesignScore(title string, skills []string, experienceTexts []string) float64 {
	score := 0.0
	if containsAny(strings.ToLower(title), []string{"designer", "ui/ux", "ux", "дизайнер", "graphic"}) {
		score += 0.40
	}
	designSkills := []string{"figma", "sketch", "adobe", "photoshop", "illustrator", "invision"}
	matched := 0
	for _, sk := range skills {
		for _, ds := range designSkills {
			if sk == ds {
				matched++
				break
			}
		}
	}
	if matched > 0 {
		score += math.Min(float64(matched)*0.10, 0.30)
	}
	designKw := []string{"design system", "wireframe", "prototype", "user research", "usability", "visual design", "макет"}
	for _, text := range experienceTexts {
		if countMatches(text, designKw) > 0 {
			score += 0.20
			break
		}
	}
	return clamp01(score)
}

// CalcDevOpsSupportScore measures DevOps capability (for non-DevOps roles this is a bonus)
func CalcDevOpsSupportScore(skills []string, experienceTexts []string) float64 {
	score := 0.0
	devopsSkills := map[string]float64{
		"docker": 0.15, "kubernetes": 0.25, "ci_cd": 0.20,
		"terraform": 0.10, "aws": 0.10, "gcp": 0.10, "azure": 0.10,
		"nginx": 0.08, "linux": 0.08, "ansible": 0.10, "prometheus": 0.10, "grafana": 0.08,
	}
	for _, sk := range skills {
		if w, ok := devopsSkills[sk]; ok {
			score += w
		}
	}
	devopsKw := []string{"deploy", "pipeline", "monitoring", "observability", "alerting", "logging", "infrastructure"}
	for _, text := range experienceTexts {
		if countMatches(text, devopsKw) > 0 {
			score += 0.15
			break
		}
	}
	return clamp01(score)
}

// CalcClientCommunicationScore measures client/stakeholder communication ability
func CalcClientCommunicationScore(experienceTexts []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"client":        0.30, "customer":    0.25, "stakeholder": 0.25,
		"requirements":  0.15, "demo":        0.15, "presentation": 0.15,
		"communication": 0.15, "negotiat":    0.10, "заказчик":     0.25,
		"клиент":        0.25, "презентац":   0.15,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcProjectManagementScore measures PM capability
func CalcProjectManagementScore(experienceTexts []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"managed project": 0.40, "coordinated": 0.25, "sprint planning": 0.20,
		"task decomposition": 0.15, "roadmap": 0.15, "resource planning": 0.15,
		"delivery":  0.15, "milestone":  0.10, "deadline":   0.10,
		"управлял":  0.30, "координир":  0.20, "планирован": 0.15,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcOwnershipScore measures ownership/responsibility
func CalcOwnershipScore(experienceTexts []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"owned":        0.35, "responsible for": 0.25, "architect":   0.20,
		"designed":     0.20, "end-to-end":      0.20, "production support": 0.15,
		"on-call":      0.10, "launched":         0.15, "built from scratch": 0.20,
		"отвечал":      0.25, "спроектировал":    0.20, "с нуля":            0.20,
		"запустил":     0.15,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcLeadershipScore measures leadership/management ability
func CalcLeadershipScore(title string, experienceTexts []string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)
	if containsAny(titleLower, []string{"lead", "head", "manager", "лид", "руководитель", "тимлид"}) {
		score += 0.45
	}
	keywords := map[string]float64{
		"led team":    0.25, "led developer": 0.25, "technical leader": 0.25,
		"mentor":      0.15, "code review":   0.10, "координировал": 0.20,
		"руководил":   0.25, "менторил":      0.15, "обучал":         0.10,
		"наставник":   0.15,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcMentoringScore measures mentoring ability
func CalcMentoringScore(experienceTexts []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"mentor":     0.40, "mentored":   0.40, "coached":   0.30,
		"onboarded":  0.25, "trained":    0.25, "teaching":  0.20,
		"менторил":   0.35, "обучал":     0.30, "наставник": 0.30,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcStartupAdaptabilityScore measures startup adaptability
func CalcStartupAdaptabilityScore(experienceTexts []string, companyCategories []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"startup": 0.35, "mvp": 0.25, "prototype": 0.20,
		"wore many hats": 0.20, "small team": 0.15, "from scratch": 0.20,
		"стартап": 0.30, "с нуля": 0.20, "прототип": 0.20,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcEngineeringEnvironmentScore measures quality of engineering practices
func CalcEngineeringEnvironmentScore(experienceTexts []string, companyEngineeringScores []float64) float64 {
	score := 0.0
	keywords := map[string]float64{
		"high-load":     0.20, "distributed":  0.20, "microservice": 0.15,
		"event-driven":  0.10, "scale":        0.10, "performance":  0.10,
		"observability": 0.10, "monitoring":   0.08, "code review":  0.08,
		"incident":      0.10, "production":   0.08, "высоконагруж": 0.20,
		"распределённ":  0.15, "микросервис":  0.15,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	// Add contribution from company engineering scores
	if len(companyEngineeringScores) > 0 {
		maxEng := 0.0
		for _, es := range companyEngineeringScores {
			if es > maxEng {
				maxEng = es
			}
		}
		score += maxEng * 0.20
	}
	return clamp01(score)
}

// CalcProjectComplexityScore measures project complexity
func CalcProjectComplexityScore(experienceTexts []string) float64 {
	score := 0.0
	keywords := map[string]float64{
		"payment":       0.20, "checkout":     0.15, "financial": 0.15,
		"distributed":   0.20, "microservice": 0.15, "high-load": 0.20,
		"scale":         0.15, "integration":  0.10, "architect":  0.15,
		"production":    0.10, "concurren":    0.15, "real-time":  0.15,
		"платёж":        0.20, "высоконагруж": 0.20, "интеграц":   0.10,
	}
	combined := strings.ToLower(strings.Join(experienceTexts, " "))
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcCompetitionScore measures competition/award achievements
func CalcCompetitionScore(achievements string, certifications string) float64 {
	score := 0.0
	combined := strings.ToLower(achievements + " " + certifications)
	keywords := map[string]float64{
		"icpc":       0.50, "olympiad":    0.40, "hackathon": 0.25,
		"winner":     0.30, "first place": 0.35, "gold medal": 0.35,
		"finalist":   0.20, "top":         0.15, "award":      0.15,
		"олимпиада":  0.40, "победитель":  0.30, "хакатон":    0.25,
		"призёр":     0.25, "1 место":     0.35,
	}
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcOpenSourceScore measures open source contributions
func CalcOpenSourceScore(experienceTexts []string, achievements string) float64 {
	score := 0.0
	combined := strings.ToLower(strings.Join(append(experienceTexts, achievements), " "))
	keywords := map[string]float64{
		"open source": 0.35, "open-source": 0.35, "github": 0.15,
		"contributor": 0.25, "maintainer":  0.35, "pull request": 0.15,
		"oss":         0.25,
	}
	for kw, w := range keywords {
		if strings.Contains(combined, kw) {
			score += w
		}
	}
	return clamp01(score)
}

// CalcGrowthTrajectoryScore measures career growth
func CalcGrowthTrajectoryScore(positions []string, companyPrestigeScores []float64) float64 {
	score := 0.0

	// Check title progression (simple heuristic)
	if len(positions) >= 2 {
		seniorityLevels := map[string]int{
			"intern": 0, "junior": 1, "middle": 2, "senior": 3, "lead": 4, "head": 5, "principal": 6,
		}
		firstLevel := -1
		lastLevel := -1
		for i, pos := range positions {
			lower := strings.ToLower(pos)
			for kw, level := range seniorityLevels {
				if strings.Contains(lower, kw) {
					if i == 0 || firstLevel == -1 {
						firstLevel = level
					}
					lastLevel = level
				}
			}
		}
		if lastLevel > firstLevel && firstLevel >= 0 {
			score += math.Min(float64(lastLevel-firstLevel)*0.15, 0.35)
		}
	}

	// Company quality improvement
	if len(companyPrestigeScores) >= 2 {
		first := companyPrestigeScores[len(companyPrestigeScores)-1] // oldest
		last := companyPrestigeScores[0]                             // newest
		if last > first {
			score += math.Min((last-first)*0.5, 0.25)
		}
	}

	// Just having multiple roles is a positive signal
	if len(positions) >= 3 {
		score += 0.15
	}

	return clamp01(score)
}

// CompanyPrestigeInput holds data for one company experience
type CompanyPrestigeInput struct {
	PrestigeScore    float64
	EngineeringScore float64
	HiringBarScore   float64
	ScaleScore       float64
	DurationMonths   int
	IsRecent         bool
	IsCurrent        bool
	IsInternship     bool
}

// CalcCompanyPrestigeScore aggregates company prestige from experience
func CalcCompanyPrestigeScore(companies []CompanyPrestigeInput) float64 {
	if len(companies) == 0 {
		return 0
	}

	totalWeight := 0.0
	totalScore := 0.0

	for i, c := range companies {
		component := 0.40*c.PrestigeScore + 0.30*c.EngineeringScore + 0.20*c.HiringBarScore + 0.10*c.ScaleScore

		// Recency weight
		recencyWeight := 0.5
		if c.IsCurrent || i == 0 {
			recencyWeight = 1.0
		} else if i == 1 {
			recencyWeight = 0.8
		}

		// Duration weight
		durationWeight := 0.5
		if c.DurationMonths >= 12 {
			durationWeight = 1.0
		} else if c.DurationMonths >= 3 {
			durationWeight = 0.8
		}

		// Type weight
		typeWeight := 1.0
		if c.IsInternship {
			typeWeight = 0.75
		}

		weight := recencyWeight * durationWeight * typeWeight
		totalScore += component * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}
	return clamp01(totalScore / totalWeight)
}

// CalcEducationQualityScore scores education quality
func CalcEducationQualityScore(institutionScores []float64, fields []string, achievements string) float64 {
	score := 0.0

	// Best institution
	if len(institutionScores) > 0 {
		maxScore := 0.0
		for _, s := range institutionScores {
			if s > maxScore {
				maxScore = s
			}
		}
		score += maxScore * 0.50
	}

	// Relevant field
	relevantFields := []string{"computer", "software", "информатик", "программ", "математик", "math", "engineering", "инженер"}
	for _, field := range fields {
		if containsAny(field, relevantFields) {
			score += 0.20
			break
		}
	}

	// Scholarship/honors
	achievementsLower := strings.ToLower(achievements)
	if containsAny(achievementsLower, []string{"scholarship", "honors", "стипенди", "отлични", "gpa 3.", "gpa 4."}) {
		score += 0.15
	}

	return clamp01(score)
}

// CalcInternshipQualityScore scores internship quality
func CalcInternshipQualityScore(internships []CompanyPrestigeInput, experienceTexts []string) float64 {
	if len(internships) == 0 {
		return 0
	}
	score := 0.0
	// Best internship company
	maxPrestige := 0.0
	for _, intern := range internships {
		p := 0.40*intern.PrestigeScore + 0.30*intern.EngineeringScore + 0.20*intern.HiringBarScore + 0.10*intern.ScaleScore
		if p > maxPrestige {
			maxPrestige = p
		}
	}
	score += maxPrestige * 0.45

	// Duration
	for _, intern := range internships {
		if intern.DurationMonths >= 3 {
			score += 0.15
			break
		}
	}

	// Project work evidence
	for _, text := range experienceTexts {
		if containsAny(text, []string{"shipped", "launched", "delivered", "запустил", "реализовал"}) {
			score += 0.20
			break
		}
	}

	return clamp01(score)
}

// CalcOverallStrength computes the overall strength aggregated score
func CalcOverallStrength(companyPrestige, engEnvironment, projectComplexity, ownership, leadership, educationQuality, internshipQuality, competition, openSource, growthTrajectory float64) float64 {
	return clamp01(
		0.18*companyPrestige +
			0.18*engEnvironment +
			0.16*projectComplexity +
			0.12*ownership +
			0.08*leadership +
			0.08*educationQuality +
			0.08*internshipQuality +
			0.06*competition +
			0.03*openSource +
			0.03*growthTrajectory,
	)
}

// CalcBackendStrength computes backend-specific strength
func CalcBackendStrength(backend, engEnvironment, projectComplexity, ownership, devopsSupport, companyPrestige, internshipQuality, clientComm, projectMgmt float64) float64 {
	return clamp01(
		0.25*backend +
			0.15*engEnvironment +
			0.15*projectComplexity +
			0.12*ownership +
			0.10*devopsSupport +
			0.10*companyPrestige +
			0.05*internshipQuality +
			0.04*clientComm +
			0.04*projectMgmt,
	)
}

// CalcFrontendStrength computes frontend-specific strength
func CalcFrontendStrength(frontend, projectComplexity, ownership, clientComm, designAwareness, companyPrestige float64) float64 {
	return clamp01(
		0.30*frontend +
			0.15*projectComplexity +
			0.15*ownership +
			0.15*clientComm +
			0.10*designAwareness +
			0.15*companyPrestige,
	)
}

// CalcDataStrength computes data-specific strength
func CalcDataStrength(data, projectComplexity, engEnvironment, ownership, companyPrestige float64) float64 {
	return clamp01(
		0.30*data +
			0.20*projectComplexity +
			0.20*engEnvironment +
			0.15*ownership +
			0.15*companyPrestige,
	)
}
