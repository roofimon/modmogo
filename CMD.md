# Useful Command Line

## MongoDB

### Open Mongo Shell

```bash
docker compose exec mongodb mongosh modmono
```

### List Products

```javascript
// list all products (any status)
db.products.find().pretty()

// only active (current state - no successors)
db.products.aggregate([
  { $lookup: { from: "products", localField: "_id", foreignField: "original_id", as: "_successors" } },
  { $match: { "_successors": { $size: 0 }, "status": "active" } }
])

// only deactivated
db.products.aggregate([
  { $lookup: { from: "products", localField: "_id", foreignField: "original_id", as: "_successors" } },
  { $match: { "_successors": { $size: 0 }, "status": "deactivated" } }
])

// find by id
db.products.findOne({ _id: ObjectId("69e5fb386567c30f46743ef7") })
```

### Other Collections

```javascript
db.customers.find().pretty()
db.orders.find().pretty()
```

### Patch: Deactivate Current Active Products with SKU Prefix SEED

```javascript
db.products.aggregate([
  { $lookup: { from: "products", localField: "_id", foreignField: "original_id", as: "_successors" } },
  { $match: { "_successors": { $size: 0 }, "status": "active", "sku": { $regex: "^SEED" } } }
]).forEach(p => {
  db.products.insertOne({
    _id: new ObjectId(),
    sku: p.sku,
    name: p.name,
    price: p.price,
    status: "deactivated",
    original_id: p._id,
    created_at: p.created_at,
    deactivated_at: new Date()
  })
})
```