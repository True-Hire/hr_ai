package main

type TechData struct {
	Name           string
	SubCategoryKeys []string
}

var Technologies = []TechData{
	// --- IT & DIGITAL ---
	{Name: "Golang", SubCategoryKeys: []string{"Information Technology (IT):Backend Development (Go, Python, PHP)"}},
	{Name: "Python", SubCategoryKeys: []string{"Information Technology (IT):Backend Development (Go, Python, PHP)", "Information Technology (IT):Data Engineering & Analytics"}},
	{Name: "React.js / Vue.js", SubCategoryKeys: []string{"Information Technology (IT):Frontend Development (React, Vue)"}},
	{Name: "Flutter / Swift / Kotlin", SubCategoryKeys: []string{"Information Technology (IT):Mobile Development (Flutter, iOS, Android)"}},
	{Name: "Docker & Kubernetes", SubCategoryKeys: []string{"Information Technology (IT):System Administration & Networking", "Information Technology (IT):Backend Development (Go, Python, PHP)"}},

	// --- DESIGN & CREATIVE ---
	{Name: "Figma", SubCategoryKeys: []string{"Information Technology (IT):UI/UX Design", "Marketing and Advertising:SMM Specialist", "Information Technology (IT):Graphic Design & Motion"}},
	{Name: "Adobe Photoshop & Illustrator", SubCategoryKeys: []string{"Information Technology (IT):Graphic Design & Motion", "Marketing and Advertising:SMM Specialist", "Textile and Light Industry:Designer (Modeler)"}},
	{Name: "CorelDRAW", SubCategoryKeys: []string{"Information Technology (IT):Graphic Design & Motion", "Textile and Light Industry:Designer (Modeler)"}},
	{Name: "After Effects / Premiere Pro", SubCategoryKeys: []string{"Information Technology (IT):Graphic Design & Motion", "Marketing and Advertising:Digital Marketer"}},

	// --- FINANCE & ACCOUNTING ---
	{Name: "1C:Buxgalteriya 8.3", SubCategoryKeys: []string{"Banking and Finance:Accounting (Buxgalteriya)", "Information Technology (IT):1C Programming & Integration"}},
	{Name: "Microsoft Excel (VBA/Macros)", SubCategoryKeys: []string{"Banking and Finance:Accounting (Buxgalteriya)", "Banking and Finance:Financial Analysis", "Information Technology (IT):Data Engineering & Analytics"}},
	{Name: "IFRS (MSFO)", SubCategoryKeys: []string{"Banking and Finance:Internal Audit", "Banking and Finance:Financial Analysis"}},
	{Name: "SAP / Oracle Finance", SubCategoryKeys: []string{"Banking and Finance:Financial Analysis", "Banking and Finance:Internal Audit"}},
	{Name: "Tax Reporting Tools", SubCategoryKeys: []string{"Banking and Finance:Accounting (Buxgalteriya)"}},

	// --- MARKETING & SALES ---
	{Name: "Google Ads / SEO Tools", SubCategoryKeys: []string{"Marketing and Advertising:Digital Marketer", "Marketing and Advertising:Targetologist"}},
	{Name: "Facebook & Instagram Ads", SubCategoryKeys: []string{"Marketing and Advertising:Targetologist", "Marketing and Advertising:SMM Specialist"}},
	{Name: "Bitrix24 / Salesforce (CRM)", SubCategoryKeys: []string{"Sales and Retail:Key Account Manager", "Sales and Retail:Store Manager (Do'kon mudiri)", "HR and Administration:Office Manager"}},
	{Name: "Copywriting / Storytelling", SubCategoryKeys: []string{"Marketing and Advertising:Copywriter", "Marketing and Advertising:SMM Specialist"}},

	// --- CONSTRUCTION & ENGINEERING ---
	{Name: "AutoCAD", SubCategoryKeys: []string{"Construction and Engineering:Civil Engineer (Muhandis)", "Construction and Engineering:Architect (Arxitektor)", "Construction and Engineering:Interior Designer"}},
	{Name: "Revit / ArchiCAD", SubCategoryKeys: []string{"Construction and Engineering:Architect (Arxitektor)", "Construction and Engineering:Interior Designer"}},
	{Name: "3ds Max / Corona Renderer", SubCategoryKeys: []string{"Construction and Engineering:Interior Designer", "Construction and Engineering:Architect (Arxitektor)"}},
	{Name: "Scrum / Lean Construction", SubCategoryKeys: []string{"Construction and Engineering:Project Manager"}},

	// --- LOGISTICS ---
	{Name: "WMS (Warehouse Management System)", SubCategoryKeys: []string{"Logistics and Transport:Warehouse Manager (Ombor mudiri)", "Logistics and Transport:Logistics Coordinator"}},
	{Name: "ERP Systems", SubCategoryKeys: []string{"Logistics and Transport:Supply Chain Manager", "Manufacturing and Production:Plant Manager"}},
	{Name: "GPS Tracking & Fleet Management", SubCategoryKeys: []string{"Logistics and Transport:Fleet Management", "Logistics and Transport:Driver (B, C, D, E categories)"}},

	// --- TEXTILE & PRODUCTION ---
	{Name: "Gerber / Lectra (CAD for Textile)", SubCategoryKeys: []string{"Textile and Light Industry:Technologist (Texnolog)", "Textile and Light Industry:Designer (Modeler)"}},
	{Name: "Industrial Sewing Machines Operating", SubCategoryKeys: []string{"Textile and Light Industry:Tailor (Tikuvchi)"}},
	{Name: "ISO Quality Standards", SubCategoryKeys: []string{"Manufacturing and Production:Quality Assurance (OTK)", "Manufacturing and Production:Plant Manager"}},
}
