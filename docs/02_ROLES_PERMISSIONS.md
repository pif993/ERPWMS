# Roles & Permissions

## Roles
- Admin: full access
- Supervisor: stock read/write, order allocate, task management
- Operator: stock read, stock move, task execution
- Viewer: read-only

## Permission matrix
- `wms.stock.read`: Admin, Supervisor, Operator, Viewer
- `wms.stock.move`: Admin, Supervisor, Operator
- `sales.order.create`: Admin, Supervisor
- `sales.order.allocate`: Admin, Supervisor
- `admin.roles.manage`: Admin
