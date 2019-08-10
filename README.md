# effe

Effe is a simple and experimental parser and evaluator for a fragment of excel's formula language.

Other projects mainly handle tokenization (github.com/xuri/efp), and push operator precedence and formula evaluation into a more comprehensive package for working with .xslx data. Some (https://github.com/handsontable/formula-parser/) handle evaluation, but using floating point math, which will break down when used for financial calculations.

Effe's goal is to support arithmetic using excel formulas, with enough precision to be appropriate for financial calculations.

Non-goals: presentation, compatability in formatting, full feature support for spreadsheet functionality in excel (eg: named tables), efficient regeneration/recalculation on input change.
