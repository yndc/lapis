# Lapis 

Lapis is a small library to build data access layers with batching and multilayer caching support. Inspired by [Facebook's dataloader](https://github.com/graphql/dataloader).

Benefits:
- Simple and consistent data loading APIs
- Reducing requests to backends through request batching
- Multi-layer caching
- Extensible through hooks

## Architecture

Repository is the component that encapsulate the data resolving layers. It is comprised of a batcher (dataloader) and layers of data resolvers exposed through a simple and consistent API.

To create a repository, use `lapis.New(config lapis.Config)`, see the [config](config) documentation to see the available configurations.

```golang
// Load an user by it's ID
user, err := userRepository.Load(1)

// Load multiple records in one call
users, errors := userRepository.LoadAll([]int{1, 2, 3})
```

### Batcher 

Data load request will be batched together, reducing backend calls and allows efficient batch loads. Batcher are included by default, although it is possible to disable them through repository configuration.

### Layer 

Layers are the data providers. They are responsible to fetch the data requested by an array of keys. 

1. If a data for a particular key is not available in a layer, the key will be passed to the next layer to be loaded. 

2. After every successful load, all layers before it will be primed with the fetched data.  

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