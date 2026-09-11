package funding

// NGI — The Landscape Governance Initiative (proposal 11 Sep 2026).
//
// What it is: a Central-African vehicle (CAR pilot, later all Africa Keystone
// Partnership signatories) that trains and equips a two-person Landscape
// Oversight Unit (LOU) inside the Ministry of Environment (solar, Starlink,
// laptops, remote-sensing + control-room training incl. Project Hoto), then
// drafts a National Landscape Plan (NLP). Budget USD ~585k (Y1) + ~650k (Y2);
// after two years NGI moves to QA/advisory and the LOU is meant to run on
// biodiversity/carbon finance (e.g. GCF). Vienna is irrelevant for NGI.
//
// Vehicle: company-like, NOT a conventional NGO — but the founder is happy to
// register an NGO/association (in CAR and/or Europe) when a funder needs one.
// Scores below = suitability for NGI given that flexibility. Entries carry
// their own "[verified 2026-09-11]" prefix; sources in verification-2026-09-11/.
var ngiSeeds = []Entry{
	// ---------- The obvious door: the Keystone Partnership itself ----------
	{Key: "keystone-partnership", Name: "Africa Keystone Protected Area Partnership – secretariat / Country Champion route", URL: "https://africakeystones.org/", Proj: ProjNGI,
		Kind: "program", Track: "africa", Amount: "no open call; partnership targets USD 1.2–1.5bn/yr into 162 PAs by 2035, RWF pays ~25%", DeadlineNote: "relationship-driven; 20 government 'Country Champions' so far – check whether CAR has signed",
		Eligibility: "Governments, NGOs, funders 'collaborating' with the partnership; no formal grant window. CMP (collaborative-management partnership) support is its stated tool.",
		Note:        "[verified 2026-09-11] NGI is written for exactly this audience: the LGI 'is designed to serve all signatories of the Keystone Partnership'. The Partnership's founding partners are RWF, WCS, AP, FZS; CI/AWF/ICCF/WWF collaborate. The government-side capacity gap NGI fills (ministry cannot cross-check NGO managers) is the weak flank of the CMP model they promote – pitch NGI as the missing 'government readiness' component. Route: RWF programme staff + ICCF Group (they run the parliamentary/ministerial side) + AP (Chinko). Ask for CAR to become a Country Champion with NGI as the readiness partner.",
		Score:       80, Why: "The proposal's own target audience; warm via AP/Chinko; no call to wait for – go talk."},
	{Key: "gef9-rwf-keystone-match", Name: "GEF-9 × Rob Walton Foundation – matched funding for Keystone PAs (via CAR GEF Operational Focal Point)", URL: "https://www.thegef.org/newsroom/news/africa-keystone-protected-area-partnership-welcomes-new-gef-rob-walton-foundation", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "RWF matches up to USD 50m (more if demand) of GEF-9 STAR money that governments direct to NGOs working in Keystone PAs; GEF-9 = USD 3.9bn, Jul 2026–Jun 2030; STAR fully flexible across BD/CC/LD", DeadlineNote: "GEF-9 started 1 Jul 2026; countries programme STAR via their Operational Focal Point (OFP) – CAR's OFP sits in the environment ministry; CI offers technical help to governments",
		Eligibility: "Money flows government → GEF agency (UNDP/UNEP/WWF/CI/IUCN…) → executing partner. Executing partner can be an NGO or a company/consultant contracted by the agency. CAR is an LDC with a STAR floor allocation.",
		Note:        "[verified 2026-09-11] Announced 31 May 2026 (GEF Assembly, Samarkand). This is the single most relevant money line for NGI: the LOU is precisely 'government oversight of NGOs managing Keystone PAs', and the OFP who decides how CAR's STAR envelope is programmed is the ministry NGI wants to embed in. Play: (1) MoU with the ministry, (2) get NGI/LOU written into CAR's GEF-9 country programming as the government-capacity component of a Keystone PA project (Chinko, Dzanga-Sangha, Bamingui-Bangoran), (3) executed via a GEF agency (CI is offering; UNDP CAR is the incumbent). 12–24 months to first disbursement – start now.",
		Score:       70, Why: "Big, structurally aligned, CAR eligible; slow and requires the ministry to carry it."},
	{Key: "rwf-direct-ngi", Name: "Rob Walton Foundation – direct approach for NGI (government-readiness component of Keystone)", URL: "https://robwaltonfoundation.org/", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "invitation-only; RWF committed ~25% of USD 1.2bn/yr Keystone target and USD 50m GEF match", DeadlineNote: "no call – relationship-based; founder already in contact re CAR parks",
		Eligibility: "Invited partners (AP, WCS, FZS, CI, AWF, ICCF…). Would need an entity to receive funds – NGO or company both seen in their portfolio (ICCF Group is a 501c3; consultancies are paid via partners).",
		Note:        "[verified 2026-09-11] Existing radar entry 'rob-walton-foundation' was dropped for the Palantir venture (no route for a tech vendor). For NGI the calculus flips: RWF funds the Keystone Partnership whose CMP model creates the very oversight gap NGI closes. Ask for a 2-year pilot grant (USD 1.2m total) or for RWF to place NGI inside an AP/CI grant as the government-facing component. Bring: ministry MoU, LOU cost sheet (USD 6k/yr salary top-ups is a talking point), Five Megapixel effort-metric as the LOU's reporting standard.",
		Score:       65, Why: "Only philanthropy that already pays for this exact problem; needs a champion inside AP/CI/ICCF."},

	// ---------- Public conservation money that reaches CAR ----------
	{Key: "eu-delegation-car", Name: "EU Delegation Bangui – NaturAfrica / ECOFAC successor grants & TA service contracts", URL: "https://international-partnerships.ec.europa.eu/policies/programming/programmes/naturafrica_en", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "NaturAfrica €310m 2021-27 continent-wide; CAR landscape actions historically €5–20m (ECOFAC 6 funded Chinko); TA service contracts €0.3–3m", DeadlineNote: "calls appear on EU Funding & Tenders portal (INTPA) and the Bangui delegation site; 2027-34 MFF programming for CAR is being written now – the moment to get 'government oversight capacity' into the country MIP",
		Eligibility: "Grants: NGOs/consortia (an NGI association or a partnership with AP/WWF/WCS works). Service contracts: companies via framework contracts or open tenders – the company-like route.",
		Note:        "[verified 2026-09-11] EU is the largest historic PA funder in CAR (ECOFAC I–VI, Dzanga-Sangha, Chinko). NaturAfrica's Central Africa priorities include 'Congo Basin forest ecosystems' and 'transhumance landscapes in Central Africa' – both squarely CAR. NaturAfrica brochure lists 'capacity-building for national agency staff' and 'management effectiveness' as special focus. Two plays: (a) join an AP/WWF-led NaturAfrica grant as the ministry-capacity partner, (b) position NGI (company) for the TA lot when the EU contracts institutional support to the ministry. Contact: EU Delegation CAR programme officer environment.",
		Score:       60, Why: "Largest CAR-relevant public funder with an explicit capacity-building mandate; slow, consortium/tender driven."},
	{Key: "cafi", Name: "CAFI – Central African Forest Initiative, CAR land-use planning window", URL: "https://cafi.org/countries/central-african-republic", Proj: ProjBoth,
		Kind: "grant", Track: "africa", Amount: "programme-scale (CAR Letter of Intent tens of USD m via UN agencies/WRI); occasional calls for expressions of interest", DeadlineNote: "no open EOI for CAR at verification; CAFI EOIs are posted on cafi.org and MPTF gateway",
		Eligibility: "Implementing organisations accredited to CAFI (UNDP, FAO, UNEP, WB, WRI, bilateral agencies); local NGOs/companies as executing partners or contractors.",
		Note:        "[verified 2026-09-11] CAFI's core deliverable in every partner country is a National Land Use Plan – NGI's 'National Landscape Plan' is the same document from the conservation side. The CAR CAFI framework already contains a land-use-planning component; the LOU would be the government unit that can actually read the satellite products CAFI pays for. Route: UNDP/WRI CAR CAFI programme managers; ask for NGI to run the ministry-capacity + NLP consultation component as contractor (company-like, fits 'not an NGO').",
		Score:       50, Why: "Exact thematic match (national land-use plan); indirect, via UN implementers."},
	{Key: "darwin-initiative", Name: "Darwin Initiative (UK Defra) – Main £200k–1m / Capability & Capacity", URL: "https://www.darwininitiative.org.uk/how-to-apply/main-applications/", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "Main £200k–£1m over 2–5 yrs (up to ~20 projects/round from 400+ applications); Capability & Capacity smaller single-stage; Extra £1–5m for scaling", DeadlineNote: "Round 32 closed (Main stage 1: 20 Jul 2026; C&C: 31 Aug 2026; stage 2 by invite 30 Nov 2026); Round 33 expected to open ~May 2027 with stage-1 ~Jul 2027; start Apr 2028",
		Eligibility: "Any legal entity worldwide may lead (NGO, university, private company on a not-for-profit basis); must deliver biodiversity + poverty outcomes in an ODA country; Round 32 targets 13 priority hotspots/35 countries – check CAR/Congo Basin is on the list; partner organisations unlimited.",
		Note:        "[verified 2026-09-11] Governance and 'systems-level change' were explicitly emphasised in Round 32. A Main project 'Landscape Oversight Unit + National Landscape Plan for CAR' with AP/Chinko and the ministry as partners is a textbook Darwin proposal; the poverty leg comes from transhumance/land-use conflict. Lead can be a UK Ltd or the CAR entity; UK lead helps. Start partnering conversations now for the mid-2027 window.",
		Score:       55, Why: "Right size, right theme, company or NGO can lead; next window mid-2027."},
	{Key: "ffem-ppi", Name: "FFEM / IUCN France – Programme de Petites Initiatives (PPI 6, 2025-2030)", URL: "https://www.programmeppi.org/aap/", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "avg €30k, up to €50k standard; PPI 7 'Solutions fondées sur la Nature' call went to €100k; ~80 projects via 8 calls 2025-30", DeadlineNote: "7th call closed 6 Mar 2026; next call expected ~Q4 2026/Q1 2027 (roughly one per semester)",
		Eligibility: "Local civil-society organisations registered in West/Central Africa – CAR explicitly eligible; young/emerging CSOs favoured; francophone; not companies.",
		Note:        "[verified 2026-09-11] Only works with an NGI association registered in CAR (founder is willing). Funds 'gestion des aires protégées' and CSO influence on environmental policy – the LOU/NLP advocacy angle fits their 'renforcement de la capacité d'influence' objective. Small but fast, and IUCN-France backing is a credibility asset with the ministry. Could fund Project Hoto or the LOU training module as a first piece.",
		Score:       45, Why: "CAR-eligible, francophone, small; needs CAR association."},
	{Key: "gef-sgp-car", Name: "GEF Small Grants Programme (UNDP CAR) – for a CAR-registered NGI association", URL: "https://sgp.undp.org/", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "up to USD 50k (strategic projects USD 150k)", DeadlineNote: "country calls set by UNDP Bangui national coordinator; verify current cycle",
		Eligibility: "Local registered CSOs/CBOs in CAR.",
		Note:        "[verified 2026-09-11] Re-scored (was dropped for the Palantir venture) because NGI can register a CAR association. Realistic for Project Hoto (urban PA control-room training site) rather than the ministry unit. Talk to the UNDP CAR SGP coordinator once the association exists.",
		Score:       35, Why: "Small; needs CAR association; good for Hoto."},
	{Key: "iki-medium-grants", Name: "IKI Medium Grants (German BMUKN) – €300k–800k", URL: "https://www.international-climate-initiative.com/en/find-funding/", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "€300k–800k per project (Medium); Small Grants €60–200k for local orgs; Large Grants €5–20m consortia", DeadlineNote: "2025/26 round: Small closed 15 Jan 2026, Medium 20 Jan 2026, Large 17 Feb 2026; next round expected ~Nov 2026–Feb 2027",
		Eligibility: "Non-profit organisations (or companies running a non-profit project) with ≥3 years of operations and audited turnover (Small: €60k–500k avg revenue) – a brand-new NGI entity is NOT eligible to lead; partner role under an established NGO possible.",
		Note:        "[verified 2026-09-11] Thematically perfect (conservation of natural carbon sinks, biodiversity, governance) and CAR is ODA-eligible – but the 3-year track record rule blocks a new vehicle. Route: sub-grantee under WWF Germany/FZS/GIZ-style lead. Re-check in 2029 as lead.",
		Score:       30, Why: "Blocked by 3-year history rule; partner-only for now."},
	{Key: "gcf-readiness-car", Name: "Green Climate Fund – Readiness & Preparatory Support (CAR NDA)", URL: "https://www.greenclimate.fund/readiness", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "up to USD 1m per country per year for NDA/institutional capacity; delivered via accredited delivery partners", DeadlineNote: "rolling; requests submitted by the country's National Designated Authority",
		Eligibility: "Requested by the NDA (in CAR the environment ministry); implemented by a GCF delivery partner (UNDP, UNEP, GIZ, FAO…) which may sub-contract.",
		Note:        "[verified 2026-09-11] The proposal names GCF as the long-run funder of the LOU. Readiness money is the plausible on-ramp: it exists to build exactly this kind of ministerial analytic capacity. Requires the minister to request it and a delivery partner to host NGI as contractor. Long lead time; pair with the GEF-9 play so the ministry asks for both at once.",
		Score:       40, Why: "Named in the proposal; government-triggered; slow."},
	{Key: "uk-blf-congo-basin", Name: "UK Biodiverse Landscapes Fund – Western Congo Basin landscape (partner role)", URL: "https://www.gov.uk/government/publications/biodiverse-landscapes-fund", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "£100m across 6 landscapes 2022-29; implementing consortia already contracted", DeadlineNote: "no open call; consortia contracted through 2029",
		Eligibility: "Sub-contract/partner of the contracted consortium (Western Congo Basin lead: verify – FFI/WWF/WCS consortium).",
		Note:        "[verified 2026-09-11] Verify whether CAR's Sangha/Dzanga corner is inside the Western Congo Basin landscape; if so, the consortium has a governance budget line and NGI's LOU is a plausible add-on. Otherwise skip.",
		Score:       20, Why: "Partner-only; geography to verify."},
	{Key: "usfws-central-africa", Name: "US Fish & Wildlife Service – Central Africa / Combating Wildlife Trafficking grants", URL: "https://www.fws.gov/program/international-affairs", Proj: ProjNGI,
		Kind: "grant", Track: "africa", Amount: "historically USD 100–500k per award (NOFOs on grants.gov)", DeadlineNote: "US international conservation NOFOs sharply reduced since 2025 – verify whether any FY2027 Africa NOFO exists",
		Eligibility: "Any organisation incl. foreign NGOs and for-profits (no fee) – company-friendly when it runs.",
		Note:        "[verified 2026-09-11] Pre-2025 the natural US door for CAR PA work (funded Chinko-type law enforcement). Status uncertain; check grants.gov quarterly, do not plan on it.",
		Score:       20, Why: "Company-eligible but programme may be dormant."},

	// ---------- Founder-led philanthropy that is structure-agnostic (company OK) ----------
	{Key: "mulago-henry-arnhold", Name: "Mulago Foundation – Henry Arnhold Fellowship (conservation/climate) – USD 100k unrestricted", URL: "https://www.mulagofoundation.org/our-fellowships", Proj: ProjBoth,
		Kind: "grant", Track: "prize", Amount: "USD 100k unrestricted + two week-long design retreats (California) + year-round coaching; ~60% of fellows enter the long-term portfolio", DeadlineNote: "applications reviewed by two staff, then up to three 'learning conversations'; third-party listings cite a 31 Oct 2026 deadline – verify on mulagofoundation.org",
		Eligibility: "Founder/leader of an organisation with a scalable solution; explicitly agnostic for-profit vs non-profit; ~20 fellows/yr from 1000+ applicants; Henry Arnhold track = conservation & climate.",
		Note:        "[verified 2026-09-11] The best structural fit for 'company-like, not a conventional NGO': Mulago funds whichever structure gets to scale. NGI's scaling story (CAR pilot → 33 Keystone countries, LOU funded by GEF/GCF after year 2) is exactly their 'design for scale' language. Founder profile (12 yrs African PAs, Chinko co-founder) is the fellow archetype. Apply with NGI as the vehicle.",
		Score:       65, Why: "Structure-agnostic USD 100k for founder-led scalable models; NGI fits the brief."},
	{Key: "fondation-segre", Name: "Fondation Segré (Geneva) – biodiversity & PA management", URL: "https://www.fondationsegre.org/", Proj: ProjNGI,
		Kind: "grant", Track: "conservation", Amount: "typically CHF 100k–1m multi-year to conservation organisations; co-funds AP/WCS/ZSL landscapes", DeadlineNote: "no open call – proposals by invitation/introduction; board decisions ~2×/yr",
		Eligibility: "Non-profit conservation organisations with field track record; species and PA-management focus.",
		Note:        "[verified 2026-09-11] Segré is one of the few European family foundations writing six-figure cheques into African PA management (incl. Central Africa). They fund partners they know – intro via AP or WCS. Needs an NGO vehicle. Pitch the LOU as the government-side complement to the AP grants they already make.",
		Score:       40, Why: "Right money, right theme; invitation-only, NGO vehicle."},
	{Key: "fondation-prince-albert", Name: "Fondation Prince Albert II de Monaco – biodiversity (Africa)", URL: "https://www.fpa2.org/en/projects/submit-a-project", Proj: ProjNGI,
		Kind: "grant", Track: "conservation", Amount: "typically €50k–300k; Africa is a priority region", DeadlineNote: "online project submission form, rolling; committee reviews several times a year",
		Eligibility: "Non-profit organisations only (associations, foundations, NGOs); no individuals, no companies.",
		Note:        "[verified 2026-09-11] Structured intake (rare among family foundations). Funds PA management and 'governance' projects in Africa. Requires the NGI association. Francophone and Central-Africa friendly.",
		Score:       40, Why: "Open intake; NGO needed; mid-size."},
	{Key: "fondation-nicolas-hulot", Name: "Fondation pour la Nature et l'Homme (ex-Fondation Nicolas Hulot)", URL: "https://www.fnh.org/", Proj: ProjNGI,
		Kind: "grant", Track: "conservation", Amount: "Génération Climat ≤€10k for 15–35-year-olds; historic 'Coups de cœur/Bourses' €1–15k; ~150 small projects/yr, mostly France", DeadlineNote: "no open international call at verification; Génération Climat rounds ~Oct (verify)",
		Eligibility: "Projects carried by a France-based structure; youth programmes age-capped 15–35; international projects historically via French partner associations.",
		Note:        "[verified 2026-09-11] Checked because the name came up: FNH today is an advocacy foundation (CESE seat, French policy) with small France-anchored grants. No instrument for a USD 585k ministry-capacity project in CAR. Only worth a shot as a personal approach to Nicolas Hulot for endorsement/visibility (he chaired the 2018 Congo Basin moves as minister) – not for money. Trash unless an introduction exists.",
		Score:       10, Why: "Not a valid funder for this; endorsement at best."},
	{Key: "synchronicity-earth", Name: "Synchronicity Earth – Congo Basin programme", URL: "https://www.synchronicityearth.org/programme/congo-basin/", Proj: ProjNGI,
		Kind: "grant", Track: "conservation", Amount: "small–mid grants (£20–100k) to under-funded local partners; flexible, multi-year", DeadlineNote: "no open call; partners found via network",
		Eligibility: "Locally-led NGOs in the Congo Basin (mainly DRC/Congo); no companies.",
		Note:        "[verified 2026-09-11] London-based re-granter with a genuine Congo Basin focus and appetite for governance/rights work. CAR is marginal in their portfolio; a CAR association led by Central Africans (the LOU staff) would be the right applicant profile, not the founder.",
		Score:       25, Why: "Small, network-driven; CAR marginal."},

	// ---------- Company-like routes: sell the service ----------
	{Key: "car-ta-tenders", Name: "TA tenders in CAR – World Bank / AfDB / UNDP / UNOPS consultancy contracts (land-use planning, PA governance)", URL: "https://projects.worldbank.org/en/projects-operations/projects-list?countrycode_exact=CF", Proj: ProjNGI,
		Kind: "procurement", Track: "africa", Amount: "consultancy contracts USD 50k–2m; published on UNDB, WB STEP, UNGM, AfDB", DeadlineNote: "rolling; monitor UNGM + WB procurement notices for CAR weekly",
		Eligibility: "Registered company (any country) with references; joint ventures with local firms common; individual consultant contracts also exist.",
		Note:        "[verified 2026-09-11] The purest 'not an NGO' path: governments in CAR buy TA for land-use planning, forest monitoring and PA governance with WB/AfDB money (e.g. WB natural-resource governance projects, CAFI-financed land-use planning). A company NGI can bid to deliver LOU set-up and NLP facilitation as a paid contract rather than a grant. Requires a registered company (CAR SARL or European) and 2–3 past-performance references (Chinko, AP consultancies count).",
		Score:       50, Why: "Company-native, repeatable; competitive and needs registration + references."},
	{Key: "earthranger-allen", Name: "EarthRanger (Allen Institute for AI) – control-room platform partnership for the LOU", URL: "https://www.earthranger.com/", Proj: ProjBoth,
		Kind: "program", Track: "program", Amount: "free platform for PAs/government agencies + onboarding support; no cash", DeadlineNote: "rolling partnership requests",
		Eligibility: "Protected-area managers, government wildlife agencies, conservation orgs.",
		Note:        "[verified 2026-09-11] The LOU's 'control room operator' training and cross-park oversight is what EarthRanger already does for AP parks – asking them to onboard a *ministry-level* view for CAR is a strong in-kind ask and a credibility signal for funders. Also relevant to Five Megapixel (telemetry effort metric).",
		Score:       45, Why: "In-kind, no cash; makes the LOU concrete."},
}

func init() { Seeds = append(Seeds, ngiSeeds...) }
