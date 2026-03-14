import re

files = [
    "/Users/dafatsq/Documents/project/dashpoint/backend/internal/handlers/report.go",
    "/Users/dafatsq/Documents/AL-Zauk/backend/internal/handlers/report.go"
]

for file_path in files:
    with open(file_path, "r") as f:
        content = f.read()

    # Sales Export CSV - Remove Cost Price (Wait, I need to check Sales Export for cost too)
    
    # Inventory Export CSV - Remove Cost Value and Cost Price
    content = content.replace('"Quantity", "Cost Price", "Sell Price", "Cost Value", "Retail Value",', '"Quantity", "Sell Price", "Retail Value",')
    # Let's use regex for multi-line lists if needed
    
    # Let's fix line 489
    
    with open(file_path, "w") as f:
        f.write(content)
