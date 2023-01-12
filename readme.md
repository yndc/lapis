# Lapis 

Lapis is a small library to build data read stores with batching and multilayer caching support. Inspired by [Facebook's dataloader](https://github.com/graphql/dataloader).

## Features 

### Streamlined data loading APIs

### Automatic multi-layer caching

### Automatic cache refresh 

TODO: Feature under development

### Batched backend calls

### Reducing redundant backend calls 

### Extensible through hooks

## Architecture

A store is an instance that handles data loading and caching for a model. Data is accessed through a key-value API such as `Load(key)`, where key could be a simple identifier or a query object. A store is comprised of a batcher (dataloader) and layers of data resolvers.

![image](https://user-images.githubusercontent.com/16462328/166320037-d88f173d-9249-4229-801d-bbf4ede297e3.png)

To create a store, use `lapis.New(config lapis.Config)`, see the [config](config) documentation to see the available configurations.

```golang
// Load an user by it's ID
user, err := userStore.Load(1)

// Load multiple records in one call
users, errors := userStore.LoadAll([]int{1, 2, 3})
```

### Batcher 

Requests (calls of `Load` and `LoadAll`) will first be processed by the batcher to be optimized. The batcher has two main job to optimize requests: request batching and request deduplication.

#### Request Batching

Requests will be first collected to be batched together, reducing backend calls and enabling efficient batch loads by the underlying storage (e.g. `MGET` for redis and single SQL query with `WHERE IN`).

![image](https://user-images.githubusercontent.com/16462328/166319997-3eaadb55-a5b6-4c9d-9675-e8bf5047b7a6.png)

#### Request Deduplication

The batcher also support deduplicating on-going requests. Suppose that you have an on-going data request call for a resource with the key `X` that takes the backend 10 seconds to load. Any subsequent requests to that resource whilst the first batch is ongoing will wait for the results from the first batch instead of creating a new request. 

This could reduce the number of slow and redundant backend calls.

### Layer 

Layers are the data providers. They are responsible to fetch the data requested by an array of keys. 

1. If a data for a particular key is not available in a layer, the key will be passed to the next layer to be loaded. 

2. After every successful load, all layers before it will be primed with the fetched data.  

![image](https://user-images.githubusercontent.com/16462328/166320081-2d77b693-98f0-40cf-be7d-e851476ceefa.png)

Examples of data layers are: in-memory cache, Redis, PostgreSQL, or external API.

Data layers implement the `lapis.Layer` interface:

```golang
type Layer[TKey comparable, TValue any] interface {
	// Unique identifier for this layer used for logging and metric purposes
	Identifier() string

	// The function that will be called to load values from the given set of keys
	Get(keys []TKey) ([]TValue, []error)

	// The function that will be called for setting data to be primed that is resolved by the layers after this
	Set(keys []TKey, values []TValue)
}
```
## Data Writes 

All Lapis stores can be primed with a data. This operation will set the given data into all layers.

Data writes is designed for priming data to reduce heavy back-end calls on data updates, we do not recommend using Lapis on its own as a read and write model. 

## Best Practice

### Layers must be idempotent 

Lapis is a library designed purely for retrieving data, and hence there is no guarantee that a request will hit all of the data layers. All layers must be idempotent and only serve to return data for the given query key.

### The final layer must be the source of truth
