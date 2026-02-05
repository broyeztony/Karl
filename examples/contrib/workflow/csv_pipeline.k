// ============================================================================
// CSV DATA PROCESSING WORKFLOW
// ============================================================================
// Demonstrates a complete data pipeline for CSV file processing:
// 1. Generate CSV data
// 2. Transform data (map values, add columns)
// 3. Filter invalid records
// 4. Validate results
// 5. Report comprehensive statistics
// ============================================================================

// Import the workflow engine
let makeEngine = import "./engine.k"
let engine = makeEngine()
let execute = engine.execute
let defaultConfig = engine.defaultConfig

log("================================================================================")
log("CSV DATA PROCESSING PIPELINE")
log("================================================================================")
log("")

// ============================================================================
// COMPLETE SEQUENTIAL PIPELINE
// ============================================================================

let csvPipeline = {
    name: "CSV Processing Pipeline",
    type: "sequential",
    tasks: [
        // ========== PHASE 1: GENERATE DATA ==========
        {
            name: "Generate Sales Data",
            handler: (ctx) -> {
                log("[PHASE 1] Generating CSV data...")
                log("  → Generating sales records...")
                sleep(50)
                
                let records = for i < 50 with i = 0, data = [] {
                    data += [{
                        id: i + 1,
                        product: if i % 3 == 0 { "Widget" } else { if i % 3 == 1 { "Gadget" } else { "Doohickey" } },
                        quantity: (i % 10) + 1,
                        price: ((i % 5) + 1) * 10,
                        region: if i % 2 == 0 { "North" } else { "South" },
                    }]
                    i = i + 1
                } then data
                
                log("  ✓ Generated", records.length, "sales records")
                { success: true, data: { sales: records } }
            },
        },
        
        // ========== PHASE 2: TRANSFORM DATA ==========
        {
            name: "Enrich Sales Data",
            handler: (ctx) -> {
                log("")
                log("[PHASE 2] Transforming data...")
                log("  → Enriching with calculated fields...")
                
                let enriched = for i < ctx.data.sales.length with i = 0, result = [] {
                    let record = ctx.data.sales[i]
                    let total = record.quantity * record.price
                    
                    result += [{
                        id: record.id,
                        product: record.product,
                        quantity: record.quantity,
                        price: record.price,
                        region: record.region,
                        total_value: total,
                        high_value: total > 200,
                        category: if record.price < 30 { "Budget" } else { if record.price < 50 { "Standard" } else { "Premium" } },
                    }]
                    i = i + 1
                } then result
                
                log("  ✓ Enriched", enriched.length, "records with 3 new columns")
                { success: true, data: { enriched: enriched } }
            },
        },
        
        // ========== PHASE 3: FILTER DATA ==========
        {
            name: "Filter High-Value Sales",
            handler: (ctx) -> {
                log("")
                log("[PHASE 3] Filtering data...")
                log("  → Applying filter: high-value sales only...")
                
                let filtered = for i < ctx.data.enriched.length with i = 0, result = [] {
                    let record = ctx.data.enriched[i]
                    if record.high_value {
                        result += [record]
                    }
                    i = i + 1
                } then result
                
                let originalCount = ctx.data.enriched.length
                let filterRate = ((originalCount - filtered.length) * 100) / originalCount
                
                log("  ✓ Filtered from", originalCount, "to", filtered.length, "records")
                log("  • Filter rate:", filterRate, "%")
                
                { success: true, data: { filtered: filtered, original_count: originalCount } }
            },
        },
        
        // ========== PHASE 4: VALIDATE RESULTS ==========
        {
            name: "Validate Data Quality",
            handler: (ctx) -> {
                log("")
                log("[PHASE 4] Validating results...")
                
                // Check record counts
                let hasRecords = ctx.data.filtered.length > 0
                log("  • Record count:", if hasRecords { "✓ VALID" } else { "✗ INVALID" })
                
                // Check all records have required fields
                let allValid = for i < ctx.data.filtered.length with i = 0, valid = true {
                    let record = ctx.data.filtered[i]
                    if record.id == 0 {
                        valid = false
                        break valid
                    }
                    i = i + 1
                } then valid
                
                log("  • Data integrity:", if allValid { "✓ PASSED" } else { "✗ FAILED" })
                
                // Check filter effectiveness
                let filterWorked = ctx.data.filtered.length < ctx.data.original_count
                log("  • Filter applied:", if filterWorked { "✓ YES" } else { "✗ NO" })
                
                { success: hasRecords && allValid && filterWorked, data: ctx.data }
            },
        },
        
        // ========== PHASE 5: GENERATE STATISTICS ==========
        {
            name: "Calculate Statistics",
            handler: (ctx) -> {
                log("")
                log("[PHASE 5] Generating statistics...")
                log("  → Computing aggregate metrics...")
                
                // Calculate sales statistics
                let salesMetrics = for i < ctx.data.filtered.length with i = 0, total = 0, count = 0, max = 0, min = 999999 {
                    let record = ctx.data.filtered[i]
                    total = total + record.total_value
                    count = count + 1
                    max = if record.total_value > max { record.total_value } else { max }
                    min = if record.total_value < min { record.total_value } else { min }
                    i = i + 1
                } then { 
                    total_revenue: total,
                    order_count: count,
                    average_order: total / count,
                    max_order: max,
                    min_order: min,
                }
                
                // Count by product
                let productCounts = for i < ctx.data.filtered.length with i = 0, widgets = 0, gadgets = 0, doohickeys = 0 {
                    let record = ctx.data.filtered[i]
                    if record.product == "Widget" { widgets = widgets + 1 }
                    if record.product == "Gadget" { gadgets = gadgets + 1 }
                    if record.product == "Doohickey" { doohickeys = doohickeys + 1 }
                    i = i + 1
                } then {
                    widgets: widgets,
                    gadgets: gadgets,
                    doohickeys: doohickeys,
                }
                
                // Count by region
                let regionCounts = for i < ctx.data.filtered.length with i = 0, north = 0, south = 0 {
                    let record = ctx.data.filtered[i]
                    if record.region == "North" { north = north + 1 }
                    if record.region == "South" { south = south + 1 }
                    i = i + 1
                } then {
                    north: north,
                    south: south,
                }
                
                // Count by category
                let categoryCounts = for i < ctx.data.filtered.length with i = 0, budget = 0, standard = 0, premium = 0 {
                    let record = ctx.data.filtered[i]
                    if record.category == "Budget" { budget = budget + 1 }
                    if record.category == "Standard" { standard = standard + 1 }
                    if record.category == "Premium" { premium = premium + 1 }
                    i = i + 1
                } then {
                    budget: budget,
                    standard: standard,
                    premium: premium,
                }
                
                log("  ✓ Calculated", 4, "metric groups")
                
                { 
                    success: true, 
                    data: {
                        filtered: ctx.data.filtered,
                        original_count: ctx.data.original_count,
                        metrics: salesMetrics,
                        products: productCounts,
                        regions: regionCounts,
                        categories: categoryCounts,
                    }
                }
            },
        },
    ],
}

// Execute the pipeline
let result = execute(csvPipeline, {}, defaultConfig)

// ============================================================================
// FINAL REPORT
// ============================================================================

log("")
log("================================================================================")
log("CSV PROCESSING PIPELINE - FINAL REPORT")
log("================================================================================")
log("")

if result.success {
    let finalData = result.results[4].data
    
    log("📊 PROCESSING SUMMARY")
    log("────────────────────────────────────────────────────────────────────────────")
    log("Pipeline Status: ✓ COMPLETED SUCCESSFULLY")
    log("")
    log("Data Flow:")
    log("  1. Generated:", finalData.original_count, "sales records")
    log("  2. Enriched with 3 new fields (total_value, high_value, category)")
    log("  3. Filtered to:", finalData.filtered.length, "high-value records")
    log("  4. Validated all records")
    log("  5. Computed comprehensive statistics")
    log("")
    
    log("📈 SALES METRICS (High-Value Orders Only)")
    log("──────────────────────────────────────────────────────────────────────")
    log("Total Orders:", finalData.metrics.order_count)
    log("Total Revenue:", finalData.metrics.total_revenue)
    log("Average Order Value:", finalData.metrics.average_order)
    log("Largest Order:", finalData.metrics.max_order)
    log("Smallest Order:", finalData.metrics.min_order)
    log("")
    
    log("📦 PRODUCT BREAKDOWN")
    log("────────────────────────────────────────────────────────────────────────────")
    log("Widgets:", finalData.products.widgets, "orders")
    log("Gadgets:", finalData.products.gadgets, "orders")
    log("Doohickeys:", finalData.products.doohickeys, "orders")
    log("")
    
    log("🗺️  REGIONAL DISTRIBUTION")
    log("────────────────────────────────────────────────────────────────────────────")
    log("North Region:", finalData.regions.north, "orders")
    log("South Region:", finalData.regions.south, "orders")
    log("")
    
    log("💰 PRICE CATEGORY ANALYSIS")
    log("────────────────────────────────────────────────────────────────────────────")
    log("Budget (<$30):", finalData.categories.budget, "orders")
    log("Standard ($30-$50):", finalData.categories.standard, "orders")
    log("Premium (>$50):", finalData.categories.premium, "orders")
    log("")
    
    log("================================================================================")
    log("✓ CSV PROCESSING PIPELINE COMPLETED SUCCESSFULLY!")
    log("================================================================================")
    log("")
    log("Pipeline demonstrated:")
    log("  ✓ Sequential data processing workflow")
    log("  ✓ Dynamic data generation")
    log("  ✓ Field enrichment and transformation")
    log("  ✓ Conditional filtering logic")
    log("  ✓ Multi-level validation checks")
    log("  ✓ Comprehensive statistical analysis")
    log("")
    log("All", 5, "phases completed with full traceability!")
} else {
    log("✗ Pipeline failed - check logs above")
}
