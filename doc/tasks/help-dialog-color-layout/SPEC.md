# Feature: Help Dialog Color Palette Layout Improvement

## Overview

Fix the help dialog color palette display where the line width exceeds the dialog width, causing an ugly stretched appearance.

## Objectives

- Reduce colors per line in the color cube and grayscale sections
- Keep the display within dialog width (70 characters)
- Maintain the `number=HEX` format for user convenience

## Functional Requirements

- FR1.1: Display color cube (16-231) with 4 colors per line
- FR1.2: Display grayscale (232-255) with 4 colors per line
- FR1.3: Maintain the `number=HEX` display format (e.g., `16=#000000`)
- FR1.4: Keep standard colors (0-15) layout unchanged (8 colors per line)

## Non-Functional Requirements

- NFR1.1: All content must fit within dialog width of 70 characters

## Out of Scope

- Changes to standard colors (0-15) layout

## Interface Contract

### Input/Output Specification

No changes to input/output interfaces.

### Affected Functions

- `renderColorCube()`: Change from 6 colors per row to 4 colors per row
- `renderGrayscale()`: Change from 6 colors per row to 4 colors per row

## Test Scenarios

- [ ] Color cube displays 4 colors per line
- [ ] Grayscale displays 4 colors per line
- [ ] Dialog does not stretch horizontally
- [ ] Existing tests pass

## Success Criteria

- [ ] Color cube section shows 4 colors per line (54 lines total for 216 colors)
- [ ] Grayscale section shows 4 colors per line (6 lines total for 24 colors)
- [ ] Dialog width remains at 70 characters
- [ ] All existing unit tests pass
