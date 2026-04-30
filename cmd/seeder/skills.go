package main

type SkillData struct {
	Name            string
	SubCategoryKeys []string // Format: "MainCategory:SubCategory"
}

var Skills = []SkillData{
	// --- INFORMATION TECHNOLOGY (IT) ---
	{Name: "Data Structures & Algorithms", SubCategoryKeys: []string{"Information Technology (IT):Backend Development (Go, Python, PHP)"}},
	{Name: "RESTful API Design", SubCategoryKeys: []string{"Information Technology (IT):Backend Development (Go, Python, PHP)", "Information Technology (IT):Frontend Development (React, Vue)"}},
	{Name: "Microservices Architecture", SubCategoryKeys: []string{"Information Technology (IT):Backend Development (Go, Python, PHP)"}},
	{Name: "State Management (Redux/Zustand)", SubCategoryKeys: []string{"Information Technology (IT):Frontend Development (React, Vue)"}},
	{Name: "Responsive Design", SubCategoryKeys: []string{"Information Technology (IT):Frontend Development (React, Vue)", "Information Technology (IT):UI/UX Design"}},
	{Name: "Cross-Platform Development", SubCategoryKeys: []string{"Information Technology (IT):Mobile Development (Flutter, iOS, Android)"}},
	{Name: "Network Security & Firewalls", SubCategoryKeys: []string{"Information Technology (IT):System Administration & Networking"}},
	{Name: "Cloud Computing (AWS/Azure/GCP)", SubCategoryKeys: []string{"Information Technology (IT):System Administration & Networking", "Information Technology (IT):Backend Development (Go, Python, PHP)"}},
	{Name: "Unit Testing & Integration Testing", SubCategoryKeys: []string{"Information Technology (IT):QA & Software Testing"}},
	{Name: "Agile/Scrum Methodology", SubCategoryKeys: []string{"Information Technology (IT):IT Project Management", "HR and Administration:HR Manager / Recruiter"}},
	{Name: "User Research & Personas", SubCategoryKeys: []string{"Information Technology (IT):UI/UX Design"}},
	{Name: "Motion Graphics", SubCategoryKeys: []string{"Information Technology (IT):Graphic Design & Motion"}},

	// --- BANKING AND FINANCE ---
	{Name: "Financial Modeling", SubCategoryKeys: []string{"Banking and Finance:Financial Analysis"}},
	{Name: "Risk Assessment & Mitigation", SubCategoryKeys: []string{"Banking and Finance:Risk Management", "Banking and Finance:Internal Audit"}},
	{Name: "Tax Accounting & Law", SubCategoryKeys: []string{"Banking and Finance:Accounting (Buxgalteriya)"}},
	{Name: "Audit Planning", SubCategoryKeys: []string{"Banking and Finance:Internal Audit"}},
	{Name: "Loan Underwriting", SubCategoryKeys: []string{"Banking and Finance:Credit Specialist"}},
	{Name: "Cash Management", SubCategoryKeys: []string{"Banking and Finance:Cashier (Kassir)", "Banking and Finance:Back Office Operations"}},

	// --- MARKETING & ADVERTISING ---
	{Name: "Search Engine Optimization (SEO)", SubCategoryKeys: []string{"Marketing and Advertising:Digital Marketer"}},
	{Name: "Social Media Strategy", SubCategoryKeys: []string{"Marketing and Advertising:SMM Specialist"}},
	{Name: "Content Writing", SubCategoryKeys: []string{"Marketing and Advertising:Copywriter", "Marketing and Advertising:SMM Specialist"}},
	{Name: "Media Buying", SubCategoryKeys: []string{"Marketing and Advertising:Targetologist"}},
	{Name: "Event Planning & Execution", SubCategoryKeys: []string{"Marketing and Advertising:Event Organizer"}},

	// --- CONSTRUCTION & ENGINEERING ---
	{Name: "Blueprint Reading", SubCategoryKeys: []string{"Construction and Engineering:Civil Engineer (Muhandis)", "Construction and Engineering:Foreman (Prorab)"}},
	{Name: "Structural Analysis", SubCategoryKeys: []string{"Construction and Engineering:Civil Engineer (Muhandis)"}},
	{Name: "Interior Spatial Planning", SubCategoryKeys: []string{"Construction and Engineering:Interior Designer"}},
	{Name: "Electrical Circuit Design", SubCategoryKeys: []string{"Construction and Engineering:Electrician (Elektrik)"}},
	{Name: "TIG/MIG Welding", SubCategoryKeys: []string{"Construction and Engineering:Welder (Payvandchi)"}},

	// --- LOGISTICS & TRANSPORT ---
	{Name: "Supply Chain Optimization", SubCategoryKeys: []string{"Logistics and Transport:Supply Chain Manager"}},
	{Name: "Inventory Control", SubCategoryKeys: []string{"Logistics and Transport:Warehouse Manager (Ombor mudiri)"}},
	{Name: "Route Planning", SubCategoryKeys: []string{"Logistics and Transport:Logistics Coordinator"}},
	{Name: "Freight Forwarding Knowledge", SubCategoryKeys: []string{"Logistics and Transport:Forwarding Agent"}},

	// --- HR & ADMINISTRATION ---
	{Name: "Talent Acquisition", SubCategoryKeys: []string{"HR and Administration:HR Manager / Recruiter"}},
	{Name: "Conflict Resolution", SubCategoryKeys: []string{"HR and Administration:HR Manager / Recruiter", "HR and Administration:Personal Assistant"}},
	{Name: "Document Archiving", SubCategoryKeys: []string{"HR and Administration:Document Specialist (Kadrlar ishi)", "HR and Administration:Office Manager"}},
	{Name: "Simultaneous Interpretation", SubCategoryKeys: []string{"HR and Administration:Translator"}},

	// --- GENERAL SOFT SKILLS (Common for many) ---
	{Name: "Critical Thinking", SubCategoryKeys: []string{"Information Technology (IT):IT Project Management", "Banking and Finance:Financial Analysis", "Education and Science:Researcher"}},
	{Name: "Negotiation Skills", SubCategoryKeys: []string{"Sales and Retail:Key Account Manager", "Sales and Retail:Real Estate Agent (Rieltov)", "Logistics and Transport:Supply Chain Manager"}},
	{Name: "Public Speaking", SubCategoryKeys: []string{"Education and Science:Teacher / Tutor", "Marketing and Advertising:PR Manager"}},
}
