# Competitive strategy

**Date:** 2026-08-27
**Primary competitor:** [diagonismoi.cy](https://diagonismoi.cy/)
**Status:** Strategy draft — informs roadmap, not yet reflected in the schema or product.

This document is the answer to one question: **diagonismoi.cy already exists, is live, and
is cheap — why would anyone use Promitheies.cy instead?**

---

## 1. What diagonismoi.cy is

A **bidder tool**, not a transparency project. It sells daily tender alerts to companies that
bid on government contracts. It is already in production with real data, and it is priced to
be an easy purchase.

### Their assets (what we will not beat head-on)

| Asset | Detail |
|---|---|
| **Scale / breadth** | Cyprus **and** Greece. ~2.07M recorded award rows (Cyprus ~57.8k, Greece ~2.0M). We will not out-scale this on breadth. |
| **Price** | **€9/month + VAT.** This is roughly one quarter of our planned €39. It is the single most important competitive fact in this document. |
| **A shipped intelligence angle** | Per-authority **median winner discount**, **average competitors per tender** (~3.1 site-wide), and a **"reliability" score** that flags authorities whose pre-award estimates look written *after* the fact (estimate == final price) — a soft corruption signal. |
| **Plain-language summaries** | Greek-language human-readable summaries of each tender's requirements. |
| **Actionable ordering** | Open tenders sorted by deadline; "closing soon" is the default view a bidder wants. |
| **Distribution** | Daily email digests by category, weekly summary emails, a "what's coming / expected re-tenders" teaser. |

### Their structure (observed)

- Navigation: `/kypros`, `/ellada`, `/diagonismoi` (open tenders), `/go` (subscription), `/syndromi` (sign in). Terms / privacy / refunds in the footer. No blog, no news, no rankings content section, no API docs.
- Tender cards show **title, description, budget estimate, deadline** — and not much else on the free surface.
- An authority ranking table: awards count, winner discount %, reliability (High/Medium), bucketed by award-volume thresholds.
- Contracts from **€10,000 and up**.
- Greek only. Email-only alerts. No visible API, bulk download, or open-data positioning.

---

## 2. Where they are weak — our openings

| Gap | Why it matters to us |
|---|---|
| **Shallow entity resolution.** Flat data model, category "tags" rather than a taxonomy. | Every headline claim they make ("who wins, at what discount") is only as good as their name-matching. If `ACME LTD` / `Acme Limited` / `ΑΚΜΕ ΛΤΔ` resolve to three contractors, every contractor profile and every incumbency statistic is silently fragmented and wrong. **We already half-built the moat**: normalized `authorities` / `contractors` tables, `*_aliases` tables, `pg_trgm` indexes, an `entities` resolution package. |
| **Greek only.** | No journalists, academics, EU audience, or international suppliers. Halves the SEO surface. |
| **No API, no bulk data, no open-data stance.** | Zero civic-tech credibility, zero backlinks, no developer or researcher audience. |
| **Pure bidder positioning.** | A free public transparency dashboard earns press and organic links that a €9 SaaS never will. That entire lane is empty. |
| **Email-only alerts.** | No Telegram, WhatsApp, Slack, webhook, RSS, or iCal. |
| **One-dimensional "reliability" metric.** | The corruption-signal idea is good but crude; it can be done far more rigorously. |
| **No company-registry linkage.** | Nobody connects contractors to the Cyprus Registrar of Companies — registration number, status, incorporation date, directors, shared-director links between firms. |
| **No post-award layer.** | Amendments, cost overruns, framework call-offs, contract duration — an industry-wide blind spot, not just theirs. |
| **Single flat price tier.** | No free-registered capture tier, no team plan. |

---

## 3. Differentiation strategy — five pillars

### Pillar 1 — Win on data trust, and make it visible

Our normalized schema is the product differentiator, but only if users can *see* it.

- On every contractor / authority profile, surface: folded-in aliases, match confidence, source citations, and a link to a methodology page.
- Add **Cyprus Registrar of Companies linkage**: registration number → company status, incorporation date, directors. This gives bulletproof entity resolution *and* unlocks "companies connected to this contractor via shared directors" — analysis diagonismoi.cy structurally cannot do.
- Publish a public methodology / data-quality page. Show our work.

### Pillar 2 — Be the definitive *Cyprus* source, not a thinner CY+GR one

Depth beats breadth on home turf.

- Below-threshold contracts, direct / negotiated awards without prior publication (the largest transparency gap), contract amendments, framework agreements.
- Coverage completeness is a claim we can defend and they cannot.

### Pillar 3 — Ship the free transparency dashboard + open API + bulk data

This is the SEO / PR engine, and it is uncontested.

- Bilingual EL / EN doubles the reach.
- Public REST API (`/api/v1/awards`, `/authorities`, `/contractors`) + downloadable CSV / JSON snapshots.
- Submit to the data.gov.cy Showcases section; pitch Cyprus Mail / Politis / Phileleftheros. Every backlink compounds.

### Pillar 4 — Deeper intelligence than "median discount"

The signals EU procurement research actually uses as red flags:

| Metric | What it says |
|---|---|
| **Single-bidder rate** (per authority / sector) | The single strongest corruption indicator in the literature. |
| **Buyer–supplier concentration** (Herfindahl index) | "This authority sends 60% of its IT spend to one firm." |
| **Repeat-winner streaks** | "Won 9 of the last 11 similar contracts from this buyer." |
| **Split-contract detection** | Same buyer + supplier, many contracts sitting just under a procedure threshold. |
| **December spending spikes** | Year-end budget dumping. |
| **Time-to-award anomalies** | Unusually fast awards; `published_at → award_date` gaps. |
| **Estimate == award value** | Their reliability signal — done properly, as a distribution, not a flag. |

### Pillar 5 — Neutralise the €9 alert war

Do **not** launch alerts at €39 against their €9. The alert feature is now commoditised.

| Plan | Price | Rationale |
|---|---|---|
| **Free (registered)** | €0 | 1 saved search, daily digest. Capture the market rather than cede it. |
| **Pro** | **€19–25/mo** | Value is the intelligence layer + multi-channel alerts + API, not the alert itself. |
| **Team** | **~€79/mo** | Up to 5 users, shared watchlists. |

Compete on *"our alerts are more accurate and arrive on Telegram,"* not on a price we will lose.

---

## 4. Build list — impact × effort

### High impact / low effort
- Bilingual (EL + EN) UI.
- Aliases + match confidence + sources visible on every profile.
- Telegram alert channel.
- iCal / RSS export of open tenders by saved search.
- Single-bidder rate + concentration score — pure SQL over data we already ingest.

### High impact / medium effort
- Public REST API + bulk CSV / JSON download.
- Auto-generated weekly summary articles (SEO surface diagonismoi.cy lacks entirely).
- CPV taxonomy with EL / EN labels, replacing free-text keyword tags.

### High impact / higher effort
- Cyprus Registrar linkage + shared-director graph.
- Contract amendment / framework-agreement tracking.
- LLM extraction of eligibility criteria, bid-bond, and contract duration from notice documents.

---

## 5. Schema additions this implies

Not yet built. Listed here so the roadmap and the next migration have a target.

| Change | Purpose |
|---|---|
| `tender_bids` (tender_id, bidder / contractor_id, amount, rank) | Real bidder counts and **single-bidder rate**. TED eForms carries this. |
| `contract_amendments` (award_id, date, new_value, reason) | The post-award layer nobody has. |
| `contractors`: add `registry_number`, `registry_status`, `incorporated_on` | Registry linkage → hard entity resolution. |
| `contractor_directors` (contractor_id, person_name_normalized) | Shared-director graph. |
| `tenders`: `is_framework bool`; `tender_cpv` join table for multiple CPVs | TED notices routinely carry several CPV codes. |
| Index `tenders.procedure_type` | Flag "negotiated without prior call for competition". |
| Materialized views: `authority_stats`, `contractor_stats`, `sector_stats` | Concentration, single-bidder rate, discount distribution, time-to-award — recomputed by the daily job. |

---

## 6. What not to do

- **Do not chase Greece.** Cyprus depth beats CY+GR breadth for our audience.
- **Do not price-match at €9.** Win on accuracy and channels.
- **Do not ship alerts before the free dashboard.** The dashboard is what makes people trust the alerts.
- **Do not scrape eprocurement.gov.cy.** Open data sources only, per the PRD.

---

## 7. One-line positioning

> **diagonismoi.cy tells a company when a tender opens. Promitheies.cy tells them — and the public — who actually wins Cyprus government contracts, and whether the process was competitive.**
