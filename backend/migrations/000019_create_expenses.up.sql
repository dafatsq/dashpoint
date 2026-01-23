-- 000019_create_expenses.up.sql
-- Create expense_categories and expenses tables for tracking operating costs

-- Expense categories table
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Expenses table
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID REFERENCES expense_categories(id) ON DELETE SET NULL,
    amount NUMERIC(15, 2) NOT NULL,
    description TEXT NOT NULL,
    expense_date DATE NOT NULL,
    vendor VARCHAR(255),
    reference_number VARCHAR(100),
    notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_expenses_category_id ON expenses(category_id);
CREATE INDEX idx_expenses_expense_date ON expenses(expense_date);
CREATE INDEX idx_expenses_created_by ON expenses(created_by);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);

-- Seed default expense categories
INSERT INTO expense_categories (name, description) VALUES
    ('Rent', 'Monthly rent or lease payments'),
    ('Utilities', 'Electricity, water, gas, internet'),
    ('Salaries', 'Employee wages and salaries'),
    ('Supplies', 'Office and store supplies'),
    ('Maintenance', 'Equipment repairs and maintenance'),
    ('Transportation', 'Delivery and transportation costs'),
    ('Marketing', 'Advertising and promotional expenses'),
    ('Insurance', 'Business insurance premiums'),
    ('Taxes & Licenses', 'Business taxes, permits, and licenses'),
    ('Other', 'Miscellaneous expenses');

-- Add comments
COMMENT ON TABLE expense_categories IS 'Categories for classifying business expenses';
COMMENT ON TABLE expenses IS 'Operating expenses for the business';
