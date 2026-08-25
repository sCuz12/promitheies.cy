# Promitheies.cy — Cyprus Procurement Intelligence Platform

## Product Requirements Document (PRD)

**Version:** 1.0
**Date:** August 25, 2026
**Author:** George
**Status:** Draft

---

## 1. Executive Summary

Promitheies.cy is a procurement intelligence platform for Cyprus that transforms scattered, hard-to-access public procurement data into a clean dashboard for citizens and a paid intelligence tool for businesses. It combines three open data sources — data.gov.cy awarded contracts, EU TED (Tenders Electronic Daily), and OpenTender/OCDS — into a single, searchable, alertable product.

The platform operates on two tiers: a free transparency dashboard that serves citizens, journalists, and researchers (driving SEO and credibility), and a paid subscription tier that serves SMEs bidding on government contracts (driving revenue).

No scraping of captcha-protected sites is required. All data comes from official open-data APIs and bulk downloads.

---

## 2. Problem Statement

### For Citizens & Journalists
- Public procurement data in Cyprus is fragmented across eprocurement.gov.cy (captcha-protected), data.gov.cy (raw CSVs), and the EU TED portal (buried among 800K+ annual notices from all EU countries)
- No single place answers: "How much did my municipality spend this year?", "Which contractors keep winning?", or "Where is taxpayer money going?"
- The existing open data portal publishes the raw data but does not present it in a way that enables oversight or understanding

### For Businesses (SMEs bidding on government contracts)
- SMEs must manually check eprocurement.gov.cy daily (through captcha verification) or pay €200+/month to generic global aggregators (TendersInfo, GlobalTenders, BidDetail) that are not Cyprus-specific
- No tool tells a bidder: "This authority tends to award to the same contractor" or "Similar contracts in this category average €X" — intelligence that determines whether preparing a bid is worth the investment
- No Cyprus-focused, Greek-language, affordable alternative exists

---

## 3. Target Users

### Primary Persona: The Cyprus SME Bidder (Paid Tier)
- Small-to-medium construction, IT, professional services, or supply companies (5–50 employees)
- Regularly bid on government contracts worth €25K–€2M
- Currently check eprocurement.gov.cy manually or rely on word-of-mouth
- Decision-maker: owner, BD manager, or office manager
- Language: Greek (primary), English (secondary)
- Willingness to pay: €29–49/month for a tool that saves time and improves win rate
- Estimated addressable market: 500–2,000 companies in Cyprus

### Secondary Persona: The Citizen/Journalist (Free Tier)
- Journalists covering governance, municipal politics, or public spending
- Opposition councillors or transparency advocates
- Engaged citizens curious about how their taxes are spent
- Researchers and academics studying public procurement

---

## 4. Data Sources

### 4.1 data.gov.cy — Awarded Contracts Dataset
- **What:** Catalog of all public contracts awarded via e-Procurement since 2009
- **Fields:** Tender title, contracting authority, award date, contractor name, contract value
- **Format:** CSV download (check for CKAN API availability)
- **Update frequency:** Periodic (to be verified — likely monthly or quarterly)
- **Coverage:** All government authorities, municipalities, and public bodies
- **Limitations:** Retrospective only (awards, not open tenders); no post-award tracking (completion, delays, cost changes)

### 4.2 TED (Tenders Electronic Daily) — EU Open Data
- **What:** All public procurement notices from EU member states above EU thresholds
- **Fields:** Notice type (planning/competition/award), contracting authority, CPV codes (sector classification), estimated value, deadline, award details
- **Format:** XML bulk downloads via FTP (ftp://ted.europa.eu, credentials: guest/guest), daily editions + monthly archives
- **Update frequency:** Daily (Monday–Friday, ~2,400 notices/day across all EU)
- **Coverage:** Cyprus tenders above EU thresholds (~€140K for supplies/services, ~€5.4M for works)
- **Limitations:** Only above-threshold contracts; XML parsing required; notices in multiple EU languages

### 4.3 OpenTender / OCDS (Open Contracting Data Standard)
- **What:** Cleaned, normalized procurement data from TED and national portals, published by the Government Transparency Institute
- **Fields:** Full contracting process data in OCDS format — planning, tender, award, contract stages; lots, bids, requirements
- **Format:** JSON (JSONL), Excel, CSV — downloadable per year or all-time
- **Update frequency:** Periodic bulk updates
- **Coverage:** Cyprus dataset available alongside 34 other jurisdictions
- **License:** CC BY-NC (Attribution-NonCommercial) — free layer is fine; paid tier needs legal review on "commercial use" interpretation, may need to use TED source data directly instead

### 4.4 Future consideration: eprocurement.gov.cy
- Currently captcha-protected; not viable for automated ingestion
- If/when they expose a public API (the new 2026 platform includes a Contract Management module), this becomes the primary real-time source
- Monitor for API availability; do not build scraping solutions

---

## 5. Product Tiers

### 5.1 Free Tier — Public Transparency Dashboard

**Purpose:** SEO engine, credibility builder, PR hook. Equivalent to what Fragmata does for dam levels.

#### Features

**F1: Headline Dashboard**
- Total contracts awarded: count and € value, current year and all-time
- Monthly/yearly trend chart
- "This month: X contracts worth €Y.YM awarded" — the single headline number
- Year-over-year comparison

**F2: Authority Profiles**
- Per ministry/department/municipality page
- Total spending, number of contracts, trend over time
- Sector breakdown (infrastructure, IT, health, supplies — derived from CPV codes)
- Top contractors by value and frequency
- URL structure: `/authority/ministry-of-health`

**F3: Contractor Profiles**
- Per company page
- Total contracts won, by value and count
- Which authorities they work with
- Sector breakdown
- URL structure: `/contractor/company-name`

**F4: Sector/Category View**
- Aggregate spend by CPV category over time
- Which authorities spend the most in each sector
- Which contractors dominate each sector

**F5: Search**
- Full-text search across tender titles, authority names, contractor names
- Filter by: date range, authority, sector/CPV, value range
- Sort by: date, value, authority

**F6: Auto-Generated Weekly Summary**
- "This week in Cyprus procurement: X contracts worth €Y awarded. Largest: [title] (€Z) to [contractor] by [authority]."
- Published as a blog post / article page for SEO
- Distributed via Telegram channel and WhatsApp group

**F7: Open API**
- Public REST API for developers/researchers
- Endpoints: `/api/awards`, `/api/authorities`, `/api/contractors`
- Rate-limited, JSON responses
- API documentation page

#### Non-Goals (Free Tier)
- No post-award tracking (completion %, delays, cost overruns)
- No bid preparation tools
- No personalized alerts

---

### 5.2 Paid Tier — Procurement Intelligence for SMEs

**Pricing:** €39/month or €349/year (≈25% annual discount)
**Trial:** 14-day free trial, no credit card required

#### Features

**P1: Smart Tender Alerts**
- User configures: sectors/CPV codes of interest, value range, preferred authorities, keywords
- System sends email + Telegram notification when a matching tender is published on TED
- Alert includes: title, authority, estimated value, deadline, direct link to TED notice
- Frequency: real-time (within hours of TED publication) or daily digest
- This is the single highest-value feature — what users pay for

**P2: Win Probability Signals (Incumbency Analysis)**
- For any open tender, show: "This authority has awarded N of its last M similar contracts to [Contractor X]"
- Concentration score: how concentrated is this authority's spending in this sector?
- Derived purely from historical award data — no editorial judgment
- Helps bidders decide whether to invest time in a bid or skip a "locked-in" tender

**P3: Pricing Benchmarks**
- "Similar contracts in [CPV category] from [authority type] averaged €X over the past 3 years"
- Distribution chart: min, median, mean, max for comparable contracts
- Helps bidders price competitively

**P4: Competitor Watch**
- Follow specific contractors
- Get notified when they win a new contract
- See their full profile: which authorities, which sectors, average contract size

**P5: Bid Deadline Calendar**
- Dashboard of all open tenders matching user's sector/preferences
- Calendar view with deadlines
- Status: open, closing soon (< 7 days), closed
- Export to Google Calendar / iCal

**P6: Saved Searches & Export**
- Save search configurations
- Export results as CSV for CRM/BD team use
- Monthly PDF report: "Your sector this month" summary

---

## 6. Technical Architecture

### 6.1 Stack

| Component | Technology | Rationale |
|---|---|---|
| Backend/API | Go | Primary stack strength; excellent for scheduled jobs, data pipelines, and API serving |
| Database | PostgreSQL | Relational data with strong full-text search; JSONB for flexible TED notice fields |
| Frontend | Static site (HTML/CSS/JS) or lightweight framework | Fast, SEO-friendly, minimal maintenance |
| Email delivery | Resend or Postmark | Transactional email for alerts |
| Telegram bot | Go telegram-bot-api | Alert distribution channel |
| Hosting | VPS (Hetzner / DigitalOcean) | Low cost, European data residency |
| Search | PostgreSQL FTS (initially); Meilisearch if needed later | Avoid premature infrastructure |

### 6.2 Data Pipeline

```
┌─────────────────────────────────────────────────┐
│                  DATA SOURCES                    │
├──────────────┬────────────────┬─────────────────┤
│  data.gov.cy │   TED FTP/API  │   OpenTender    │
│  (CSV/CKAN)  │   (XML daily)  │   (JSONL bulk)  │
└──────┬───────┴───────┬────────┴────────┬────────┘
       │               │                 │
       ▼               ▼                 ▼
┌─────────────────────────────────────────────────┐
│              INGESTION LAYER (Go)                │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Awards   │  │ TED      │  │ OCDS          │  │
│  │ Importer │  │ Parser   │  │ Normalizer    │  │
│  └────┬─────┘  └────┬─────┘  └──────┬────────┘  │
│       │              │               │           │
│       ▼              ▼               ▼           │
│  ┌─────────────────────────────────────────┐     │
│  │         Deduplication & Matching        │     │
│  │   (same tender across sources → merge)  │     │
│  └────────────────┬────────────────────────┘     │
└───────────────────┼──────────────────────────────┘
                    ▼
┌─────────────────────────────────────────────────┐
│              POSTGRESQL DATABASE                 │
│                                                  │
│  Tables:                                         │
│  - authorities (id, name, type, region)           │
│  - contractors (id, name, normalized_name)        │
│  - tenders (id, title, authority_id, cpv_codes,  │
│    estimated_value, deadline, status, source)     │
│  - awards (id, tender_id, contractor_id, value,  │
│    award_date, source)                            │
│  - cpv_categories (code, description_en,          │
│    description_el)                                │
│  - user_alerts (user_id, cpv_codes, keywords,    │
│    min_value, max_value, authorities)             │
│  - users (id, email, tier, telegram_id)           │
└───────────────────┬──────────────────────────────┘
                    │
         ┌──────────┼──────────┐
         ▼          ▼          ▼
   ┌──────────┐ ┌────────┐ ┌──────────┐
   │ REST API │ │ Alert  │ │ Static   │
   │ (Go)     │ │ Engine │ │ Site Gen │
   │          │ │ (Go)   │ │          │
   └──────────┘ └────────┘ └──────────┘
```

### 6.3 Scheduled Jobs

| Job | Frequency | Source | Action |
|---|---|---|---|
| TED daily ingest | Daily 06:00 UTC | TED FTP daily XML | Parse, filter country=CY, insert new tenders/awards |
| data.gov.cy sync | Weekly (Sunday) | data.gov.cy CSV | Download, diff against existing, insert new awards |
| OpenTender bulk sync | Monthly | OCDS JSONL download | Full reconciliation, fill gaps |
| Alert dispatch | Daily 07:00 UTC + real-time | Internal DB | Match new tenders against user alert configs, send notifications |
| Weekly summary generation | Monday 06:00 UTC | Internal DB | Generate weekly summary article |
| Authority/Contractor stats | Daily 05:00 UTC | Internal DB | Recompute aggregate statistics |

### 6.4 Entity Resolution

A key technical challenge: the same authority or contractor may appear with slightly different names across sources (e.g., "Υπουργείο Υγείας" vs "Ministry of Health" vs "MOH"). The pipeline needs:

- **Authority matching:** Build a canonical authority registry; map source-specific names to canonical IDs using fuzzy matching + manual overrides table
- **Contractor matching:** Company names vary (Ltd vs Limited, typos, Greek vs English). Start with exact match + simple normalization (lowercase, strip legal suffixes), add fuzzy matching if needed
- **Tender deduplication:** Same tender may appear on TED and data.gov.cy — match by authority + title + approximate date + value

---

## 7. Information Architecture & Pages

### Public Pages (Free)

```
/                          → Dashboard (headline numbers, recent awards)
/authorities               → All authorities list
/authority/:slug           → Authority profile page
/contractors               → All contractors list (sortable by total value)
/contractor/:slug          → Contractor profile page
/sectors                   → CPV category overview
/sector/:cpv-code          → Sector detail page
/search                    → Search interface
/articles                  → Weekly summaries archive
/articles/:slug            → Individual article
/api                       → API documentation
/api/v1/awards             → Awards endpoint
/api/v1/authorities        → Authorities endpoint
/api/v1/contractors        → Contractors endpoint
/about                     → About, methodology, data sources
```

### Authenticated Pages (Paid)

```
/dashboard                 → Personalized dashboard (matching tenders, deadlines)
/alerts                    → Alert configuration
/alerts/history            → Past alerts sent
/calendar                  → Bid deadline calendar
/analysis/:tender-id       → Win probability analysis for a specific tender
/benchmarks                → Pricing benchmark tool
/watchlist                 → Followed contractors
/exports                   → Saved searches and CSV exports
/settings                  → Account, billing, notification preferences
```

---

## 8. Monetization

### Pricing

| Plan | Price | Includes |
|---|---|---|
| Free | €0 | Dashboard, search, authority/contractor profiles, API (rate-limited) |
| Pro | €39/month or €349/year | Smart alerts, win probability signals, pricing benchmarks, competitor watch, calendar, CSV export |
| Team | €89/month or €799/year (future) | Pro features for up to 5 users, shared watchlists, team alert management |

### Payment
- Stripe integration
- 14-day free trial (no credit card)
- Monthly or annual billing

### Revenue Projections (Conservative)

| Metric | Year 1 | Year 2 |
|---|---|---|
| Free users | 500–1,000 | 2,000–5,000 |
| Paid subscribers | 30–60 | 100–200 |
| MRR (at €39/mo avg) | €1,170–€2,340 | €3,900–€7,800 |
| ARR | €14K–€28K | €47K–€94K |

---

## 9. Multilingual Strategy

- **Primary language:** Greek (ελληνικά) — this is where the local SME audience is
- **Secondary language:** English — for international contractors, journalists, researchers, and SEO
- **Data language:** Tender titles and descriptions arrive in Greek from national sources and in the notice's original language from TED; store both, display based on user language preference
- **UI:** i18n from day one (Go templates with language keys); start with EL + EN

---

## 10. Launch Plan

### Phase 1: Data Pipeline & Free Dashboard (Weeks 1–4)

**Week 1–2: Data ingestion**
- Set up PostgreSQL schema
- Build Go importer for data.gov.cy CSV (historical awards)
- Build Go parser for TED XML daily feed (filter Cyprus)
- Entity resolution: authority and contractor normalization
- Verify data quality: spot-check 50 records against source

**Week 3–4: Free dashboard**
- Headline dashboard page
- Authority profile pages
- Contractor profile pages
- Search functionality
- Basic SEO: meta tags, structured data, sitemap
- Telegram channel setup
- First weekly summary article (manual, to establish format)
- Deploy to production

**Launch checklist:**
- [ ] Data pipeline running reliably for 7+ days
- [ ] 10,000+ awards in database
- [ ] 50+ authority profiles generated
- [ ] 200+ contractor profiles generated
- [ ] Search returns relevant results
- [ ] Greek + English UI complete
- [ ] Telegram channel created with first post
- [ ] Submit to data.gov.cy Showcases section

### Phase 2: Paid Tier MVP (Weeks 5–8)

**Week 5–6: Alert engine**
- User registration / authentication (email + password or magic link)
- Alert configuration UI (select CPV codes, keywords, value range)
- Alert matching engine: new TED notices → user configs → queue
- Email delivery (Resend/Postmark integration)
- Telegram bot for alert delivery

**Week 7–8: Intelligence features**
- Win probability / incumbency analysis (computed from historical data)
- Pricing benchmarks (aggregated from awards by CPV)
- Competitor watch (follow a contractor)
- Stripe integration for billing
- 14-day trial flow

**Launch checklist:**
- [ ] Alert engine sending test alerts reliably for 7+ days
- [ ] Incumbency analysis producing sensible results
- [ ] Stripe test payments working
- [ ] Trial → paid conversion flow tested
- [ ] Landing page with value proposition and pricing

### Phase 3: Growth & Iteration (Months 3–6)

- Auto-generated weekly summary articles (AI-assisted from structured data)
- Bid deadline calendar with iCal export
- CSV export for paid users
- Public API documentation and developer page
- Outreach: submit to data.gov.cy Showcases, pitch to Cyprus Mail, local business press
- WhatsApp channel
- Iterate based on user feedback: which alerts get opened, which features get used

---

## 11. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| data.gov.cy CSV format changes or breaks | Medium | Medium | Build importer with validation layer; alert on schema changes; manual fallback |
| TED FTP access changes | Low | High | TED is EU infrastructure, very stable; also available via data.europa.eu; OpenTender as tertiary fallback |
| OpenTender CC BY-NC license blocks paid tier | Medium | Medium | Use OpenTender for initial data seeding and reconciliation only; serve paid features from TED source data directly (no license restriction) |
| Low SME willingness to pay | Medium | High | Validate with 10 manual conversations before building paid tier; free tier builds audience regardless |
| Entity resolution errors (wrong contractor matched) | Medium | Medium | Start with exact match + simple normalization; surface confidence scores; allow manual corrections |
| eprocurement.gov.cy launches its own good UI | Low | High | Differentiate on intelligence layer (incumbency, benchmarks) not just data presentation; government portals historically don't iterate on UX |
| Legal pushback from contractors named in profiles | Low | Low | All data is government-published open data; profiles are factual republishing with citation, not editorial claims |

---

## 12. Success Metrics

### Phase 1 (Month 1–2)
- Data pipeline uptime > 99%
- 10,000+ awards ingested
- 100+ unique visitors/week to dashboard
- Listed on data.gov.cy Showcases
- 50+ Telegram channel subscribers

### Phase 2 (Month 3–4)
- 20+ free trial signups
- 10+ paid subscribers
- Alert open rate > 40%
- < 5 entity resolution errors reported

### Phase 3 (Month 5–6)
- 50+ paid subscribers
- MRR > €1,500
- 1,000+ unique visitors/week
- 1+ media mention (Cyprus Mail, Phileleftheros, etc.)
- NPS > 40 among paid users

---

## 13. Open Questions

1. **CKAN API on data.gov.cy:** Does data.gov.cy expose a CKAN-compatible API, or is it CSV-download only? This determines whether the awards sync is a clean API call or a file-download-and-parse job.

2. **OpenTender license for commercial use:** The CC BY-NC license on OpenTender data may restrict its use in the paid tier. Legal review needed. Fallback: use TED source data directly for paid features.

3. **TED eForms transition:** TED is migrating to the eForms data standard. Verify whether existing XML parsers need updating for the new format.

4. **CPV code coverage in data.gov.cy:** Does the national awards dataset include CPV codes, or only tender titles? CPV codes are critical for sector classification and alert matching. If missing, derive from title using keyword mapping.

5. **Pricing validation:** Is €39/month the right price point for Cyprus SMEs? Validate with 10 conversations before committing.

6. **Name:** "Promitheies" (Προμήθειες = procurements) is descriptive but generic. Consider alternatives: Diagwnismoi.cy, Symvaseis.cy, or a non-literal name.

---

## 14. Appendix

### A. CPV Code Examples (Common in Cyprus)

| CPV Code | Description |
|---|---|
| 45000000 | Construction work |
| 71000000 | Architectural, engineering, and planning services |
| 72000000 | IT services |
| 33000000 | Medical equipment and pharmaceuticals |
| 34000000 | Transport equipment |
| 79000000 | Business services (legal, consulting, security) |
| 44000000 | Construction materials |
| 50000000 | Repair and maintenance services |

### B. Competitive Landscape

| Competitor | Coverage | Price | Cyprus-Specific? | Intelligence Layer? |
|---|---|---|---|---|
| eprocurement.gov.cy | Cyprus only | Free | Yes | No (raw data, captcha) |
| TendersInfo | Global | €200+/mo | No | Basic |
| GlobalTenders | Global | €150+/mo | No | Basic |
| BidDetail | Global | Varies | No | No |
| OpenTender | EU-wide | Free | Partially | Basic analytics |
| **Promitheies.cy** | **Cyprus** | **€39/mo** | **Yes** | **Yes (incumbency, benchmarks)** |

### C. Reference Projects

- **Fragmata.info** — Cyprus dam levels dashboard (same civic-data-to-dashboard pattern)
- **OpenTender.eu** — EU procurement analytics (inspiration for intelligence features)
- **AppaltiPOP (Italy)** — Italian procurement transparency platform built on OCDS data
- **PublicBI** — EU-wide procurement dashboard on TED data