# HTMX Form Handler

Added a new HTMX-based form handler that provides:
- Bootstrap-styled form components
- HTMX-powered form submission and results updating
- Streaming results support
- Loading indicators
- Error handling
- Responsive layout

The handler is similar to the datatables handler but modernized with HTMX for better interactivity and Bootstrap for improved styling.

Added example route `/example-htmx` to the demo server to showcase the HTMX form handler functionality.

## Bug Fixes
- Fixed template error by using correct field name `ShortDescription` instead of `Description` for sections
- Improved HTMX form handling with separate routes for form and results
- Added dedicated results template for cleaner partial updates 

# HTMX Form URL State Management

Added URL state management to HTMX forms to enable browser history navigation and form state persistence through URLs.

- Implemented custom HTMX extension 'push-url-w-params' for proper URL parameter handling
- Updated form to use extension with data-push-url attribute
- Updated HandleForm to process query parameters for HTMX requests
- Added form state rehydration from URL parameters 