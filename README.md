To run project:

1. Install golang;

In project repo prompt

2. go mod tidy;
3. go mod vendor;
3. go run main.go.

Can require installation of some packages with "go get ..."

Also it is mandatory to add in root repo config file from google cloud with running Google Vision API and Video Intelligence API

To deploy:

Change parameters in dctl.sh in each app according to your credentials.

1. chmod +x dctl.sh
2. ./dctl.sh
3. Chill and see how it works