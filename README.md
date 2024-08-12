# A Salesforce CLI written in Go

A Go app that uses the Salesforce API to create, query, update and delete CRM objects like accounts and contacts. Go is a cross-platform, statically typed and compiled programming language. A dev container (development container) is a self-contained dev environment with a Docker container allowing portability and instant creation of a dev environment without manual installation.

## Lastest update

This is a new project and actively under development.

## Currently implemented functionality
1. Reads environment variables required (Consumer Key, Consumer Secret, Salesforce Url)
1. Retrieves an OAuth Access Token from Salesforce 
1. Lets users search for Accounts, Contacts, Tasks and prints results
1. Retry logic to get a new OAuth Access Token if a Salesforce call fails e.g., token has expired

## Developer Salesforce license

Register for [a complimentary developer account](https://developer.salesforce.com/signup) to use with Salesforce API testing

## dev container

I'm using a dev container so I don't have to install Go on my Mac. All I need a is a Docker daemon, which in my case is `colima` and VS Code with the dev container extension.

## Authentication

Instance URL, Consumer Key and Consumer Secret are read as environment variables which you place in `.zshrc` or `.bashrc`

```sh
# set SalesForce environment variables
export SALESFORCE_CONSUMER_KEY_1=""
export SALESFORCE_CONSUMER_SECRET_1=""
export SALESFORCE_URL_1=""
```

Retrieve the Url from the Salesforce UI, View Profile and the Url is under your profile user name.
 
Retrieve the Consumer Key and Consumer Secret from the Salesforce UI, View Setup, App Manager, Connected Apps.

The app authenticates uses these environment variables and generates an OAuth Access Token that is used for SOQL Salesforce calls.

## Multiple Salesforce deployment support

The app allows up to 2 Salesforce deployments. When the app starts, the first one entered is loaded. There is an action in the CLI to switch to another deployment if environment variables have been entered.

```sh
# set environment variables
export SALESFORCE_CONSUMER_KEY_1=""
export SALESFORCE_CONSUMER_SECRET_1=""
export SALESFORCE_URL_1=""
export SALESFORCE_CONSUMER_KEY_2=""
export SALESFORCE_CONSUMER_SECRET_2=""
export SALESFORCE_URL_2=""
```

## Resources

[Go](https://go.dev/)

[Dev Container specification](https://containers.dev/implementors/spec/)

## License

This project is licensed under the [MIT License](LICENSE)

## Contributing

### Disclaimer: Unmaintained and Untested Code

Please note that this program is not actively maintained or tested. While it may work as intended, it's possible that it will break or behave unexpectedly due to changes in dependencies, environments, or other factors.

Use this program at your own risk, and be aware that:
1. Bugs may not be fixed
1. Compatibility issues may arise
1. Security vulnerabilities may exist

If you encounter any issues or have concerns, feel free to open an issue or submit a pull request.