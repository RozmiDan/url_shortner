terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
}

provider "yandex" {
  cloud_id  = var.cloud_id
  folder_id = var.folder_id
  zone      = "ru-central1-a"
  service_account_key_file = "/home/daniel/key.json"
}

data "yandex_vpc_network" "existing_network" {
  name = "default"
}

resource "yandex_vpc_subnet" "default" {
  name           = "my-subnet"
  zone           = "ru-central1-a"
  network_id     = data.yandex_vpc_network.existing_network.id
  v4_cidr_blocks = ["10.0.0.0/24"]
}

data "yandex_compute_image" "ubuntu_image" {
  family = "ubuntu-2204-lts"
}

resource "yandex_compute_instance" "vm" {
  name        = "devops-vm"
  zone        = "ru-central1-a"
  hostname    = "devops-vm"
  platform_id = "standard-v1"

  resources {
    cores  = 4
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      type     = "network-hdd"
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.default.id
    nat       = true
  }

  metadata = {
    ssh-keys = "ubuntu:${file("/home/daniel/.ssh/id_ed25519.pub")}"

    user-data = <<EOF
#cloud-config
packages:
  - docker.io
runcmd:
  - systemctl enable docker
  - systemctl start docker
EOF
  }
}
