terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-586877430255"
    dynamodb_table = "terraform-state-lock-586877430255"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "shure-monitoring.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-cluster-endpoint"
}

provider "kubernetes" {
  host = data.aws_ssm_parameter.eks_cluster_endpoint.value
}

data "aws_ssm_parameter" "db_address" {
  name = "/env/couch-address"
}

data "aws_ssm_parameter" "db_user" {
  name = "/env/couch-username"
}

data "aws_ssm_parameter" "db_pass" {
  name = "/env/couch-password"
}

data "aws_ssm_parameter" "eventhub_address" {
  name = "/env/hub-address"
}

module "shipyard_prd" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "shure-monitoring-prd"
  image          = "docker.pkg.github.com/byuoitav/shure-monitoring-service/shure-monitoring-service-dev"
  image_version  = "a1e7745"
  container_port = 80 // doesn't actually exist in container
  repo_url       = "https://github.com/byuoitav/shure-monitoring-service"

  // optional
  image_pull_secret = "github-docker-registry"
  container_args = [
    "--db-address", data.aws_ssm_parameter.db_address.value,
    "--db-username", data.aws_ssm_parameter.db_user.value,
    "--db-password", data.aws_ssm_parameter.db_pass.value,
    "--eventhub-address", data.aws_ssm_parameter.eventhub_address.value,
  ]
  health_check = false
}
