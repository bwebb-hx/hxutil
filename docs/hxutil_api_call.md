## hxutil api call

Calls a given URI as a one-off test, and shows the response

### Synopsis

Calls a given URI as a one-off test, and shows the response.

You can optionally enter variable naems in the URI, such as ":p-id" (or :project-id), :d-id (:datastore-id), etc
which will be automatically replaced with the config variables, user, etc.

// do a GET request, with config variables (:d-id) applied to the URI and the auth flag set
hxutil api call /api/v0/datastores/:d-id/actions -a

// do a POST call with a payload, without authorization
hxutil api call /api/v0/login -m POST -b '{ "email": "user@company.com", "password": "xyz" }'

```
hxutil api call [flags]
```

### Options

```
  -a, --auth            if flag is set, config is used to get hexabase auth token to pass in authorization header.
  -b, --body string     body payload to pass when calling the API. only used for POST requests.
  -h, --help            help for call
  -m, --method string   method to use when calling the API. (default "GET")
```

### SEE ALSO

* [hxutil api](hxutil_api.md)	 - Utilities for testing Hexabase APIs

###### Auto generated by spf13/cobra on 6-Dec-2024