UF THETA TAU WEBSITE - BUILD PLAN
Features: 1. View Active Brothers, 2. View Active PNMs 3. Viewable Bro Database
 
 
PREREQUISITES 
----------------------------------------
- Agree on the schema, roles and endpoints
- Auth: brothers log in- Im thinking Google OAuth or an invite-only list
- Roles: brother, officer (rush chair, VP, etc.), admin.
- Middleware: requireLogin and requireRole(...) in the Go backend.
- What will we use for file storage
- How will we add new brothers to the database
