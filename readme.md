# Lapis 

Lapis is a small library to build data read store with batching and multilayer caching support. Inspired by [Facebook's dataloader](https://github.com/graphql/dataloader).

Benefits:
- Simple and consistent data loading APIs
- Reducing requests to backends through request batching
- Multi-layer caching
- Extensible through hooks

## Architecture

Store is the component that encapsulate the data resolving layers. It is comprised of a batcher (dataloader) and layers of data resolvers exposed through a simple and consistent API.

![image](https://user-images.githubusercontent.com/16462328/166320037-d88f173d-9249-4229-801d-bbf4ede297e3.png)

To create a store, use `lapis.New(config lapis.Config)`, see the [config](config) documentation to see the available configurations.

```golang
// Load an user by it's ID
user, err := userStore.Load(1)

// Load multiple records in one call
users, errors := userStore.LoadAll([]int{1, 2, 3})
```

### Batcher 

Requests will first be processed by the batcher to be optimized, by deduplicating requests of the same resource and collecting multiple requests to be resolved together in one batch. Batched can be enabled by providing a batcher config in the creation of `Store`

#### Request Collector

Data load requests of a store will be collected first to be batched together, reducing backend calls and allows efficient batch loads by the layers (e.g. `MGET` for redis and single SQL query with `WHERE IN`). If a store has a batcher enabled, the collector is also enabled by default. It is possible to disable the collector by using `LoadNoCollectBatch` when loading resources.

![image](https://user-images.githubusercontent.com/16462328/166319997-3eaadb55-a5b6-4c9d-9675-e8bf5047b7a6.png)

#### Request Sharing

The batcher also support deduplicating on-going requests. Suppose that you have an on-going data request call for a resource with the key `X` that takes 10 seconds to load (suppose that it's not cached yet). Any subsequent requests to that resource whilst the first batch is ongoing will wait for the results from the first batch instead of creating a new request. 

This could reduce the number of slow and redundant backend calls. Although this could affect the response time for resources that is supposed to be fast (e.g. cached). 

It is recommended to use request collector or sharing exclusively, collector is more suitable on resources that are fast to resolve with high variation of requests. Request sharing is more suitable for slow resolvers with low variation of requests.

### Layer 

Layers are the data providers. They are responsible to fetch the data requested by an array of keys. 

1. If a data for a particular key is not available in a layer, the key will be passed to the next layer to be loaded. 

2. After every successful load, all layers before it will be primed with the fetched data.  

![image](https://user-images.githubusercontent.com/16462328/166320081-2d77b693-98f0-40cf-be7d-e851476ceefa.png)

Examples of data layers are: in-memory cache, Redis, PostgreSQL, or external API.

Implementing a data layer is really easy, they implement the `lapis.Layer` interface:

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