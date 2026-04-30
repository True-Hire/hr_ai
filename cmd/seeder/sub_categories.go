package main

type SubCategoryData struct {
	MainCategory string
	Names        []string
}

var SubCategories = []SubCategoryData{
	{
		MainCategory: "Information Technology (IT)",
		Names: []string{
			"Backend Development (Go, Python, PHP)",
			"Frontend Development (React, Vue)",
			"Mobile Development (Flutter, iOS, Android)",
			"System Administration & Networking",
			"1C Programming & Integration",
			"UI/UX Design",
			"QA & Software Testing",
			"IT Project Management",
			"Data Engineering & Analytics",
			"Graphic Design & Motion",
		},
	},
	{
		MainCategory: "Banking and Finance",
		Names: []string{
			"Credit Specialist",
			"Accounting (Buxgalteriya)",
			"Cashier (Kassir)",
			"Financial Analysis",
			"Internal Audit",
			"Risk Management",
			"Back Office Operations",
			"Currency Operations",
		},
	},
	{
		MainCategory: "Textile and Light Industry",
		Names: []string{
			"Tailor (Tikuvchi)",
			"Technologist (Texnolog)",
			"Designer (Modeler)",
			"Fabric Quality Control",
			"Production Manager",
			"Cutter (Bichuvchi)",
			"Machine Operator",
		},
	},
	{
		MainCategory: "Sales and Retail",
		Names: []string{
			"Sales Consultant (Sotuvchi-konsultant)",
			"Store Manager (Do'kon mudiri)",
			"Cashier (Kassir)",
			"Merchandiser",
			"Key Account Manager",
			"Real Estate Agent (Rieltov)",
			"Pharmacy Sales (Farmatsevt)",
		},
	},
	{
		MainCategory: "Construction and Engineering",
		Names: []string{
			"Civil Engineer (Muhandis)",
			"Architect (Arxitektor)",
			"Project Manager",
			"Electrician (Elektrik)",
			"Welder (Payvandchi)",
			"Interior Designer",
			"Plumber (Santexnik)",
			"Foreman (Prorab)",
		},
	},
	{
		MainCategory: "Food and Hospitality",
		Names: []string{
			"Chef (Oshpaz)",
			"Pastry Chef (Qandolatchi)",
			"Waiter (Ofitsiant)",
			"Administrator (Restoran/Mehmonxona)",
			"Barista / Bartender",
			"Food Technologist",
			"Hotel Manager",
		},
	},
	{
		MainCategory: "Logistics and Transport",
		Names: []string{
			"Logistics Coordinator",
			"Driver (B, C, D, E categories)",
			"Warehouse Manager (Ombor mudiri)",
			"Forwarding Agent",
			"Expeditor",
			"Supply Chain Manager",
			"Courier (Kuryer)",
			"Fleet Management",
		},
	},
	{
		MainCategory: "Manufacturing and Production",
		Names: []string{
			"Plant Manager",
			"Quality Assurance (OTK)",
			"Mechanical Engineer",
			"Electrical Engineer",
			"Production Operator",
			"Safety Engineer (TB)",
		},
	},
	{
		MainCategory: "Agriculture",
		Names: []string{
			"Agronomist (Agronom)",
			"Veterinarian (Veterinar)",
			"Greenhouse Specialist",
			"Farm Manager",
			"Irrigation Specialist",
		},
	},
	{
		MainCategory: "Medicine and Healthcare",
		Names: []string{
			"General Practitioner",
			"Specialized Doctor",
			"Nurse (Hamshira)",
			"Laboratory Technician",
			"Pharmacist",
			"Medical Representative",
		},
	},
	{
		MainCategory: "Education and Science",
		Names: []string{
			"Teacher / Tutor",
			"Language Instructor (IELTS/CEFR)",
			"Preschool Teacher",
			"Academic Coordinator",
			"Researcher",
		},
	},
	{
		MainCategory: "Marketing and Advertising",
		Names: []string{
			"SMM Specialist",
			"Digital Marketer",
			"Copywriter",
			"PR Manager",
			"Targetologist",
			"Event Organizer",
		},
	},
	{
		MainCategory: "HR and Administration",
		Names: []string{
			"HR Manager / Recruiter",
			"Office Manager",
			"Personal Assistant",
			"Document Specialist (Kadrlar ishi)",
			"Translator",
		},
	},
	{
		MainCategory: "Beauty and Personal Care",
		Names: []string{
			"Hairdresser",
			"Make-up Artist",
			"Nail Technician",
			"Cosmetologist",
			"Fitness Instructor",
		},
	},
}
