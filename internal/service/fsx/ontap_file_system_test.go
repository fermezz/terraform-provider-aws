package fsx_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxOntapFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1024"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.OntapDeploymentTypeMultiAz1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_ip_address_range"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_vpc.test", "default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_subnet_id", "aws_subnet.test1", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.intercluster.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.intercluster.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxOntapFileSystem_fsxSingleAz(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemSingleAzConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.OntapDeploymentTypeSingleAz1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxOntapFileSystem_fsxAdminPassword(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemFsxAdminPasswordConfig(rName, pass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids", "fsx_admin_password"},
			},
			{
				Config: testAccOntapFileSystemFsxAdminPasswordConfig(rName, pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass2),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_endpointIPAddressRange(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemEndpointIPAddressRangeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "endpoint_ip_address_range", "198.19.255.0/24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxOntapFileSystem_diskIops(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemDiskIopsConfigurationConfig(rName, 3072),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemDiskIopsConfigurationConfig(rName, 4000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "4000"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceOntapFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOntapFileSystem_securityGroupIDs(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemSecurityGroupIds1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemSecurityGroupIds2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_routeTableIDs(t *testing.T) {
	var filesystem1 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemRouteTableConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxOntapFileSystem_tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOntapFileSystemTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem3),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemWeeklyMaintenanceStartTimeConfig(rName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemWeeklyMaintenanceStartTimeConfig(rName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_automaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemAutomaticBackupRetentionDaysConfig(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemAutomaticBackupRetentionDaysConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccOntapFileSystemAutomaticBackupRetentionDaysConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "1"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_kmsKeyID(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemKMSKeyIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccFSxOntapFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemDailyAutomaticBackupStartTimeConfig(rName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemDailyAutomaticBackupStartTimeConfig(rName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_throughputCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemThroughputCapacityConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "256"),
				),
			},
		},
	})
}

func TestAccFSxOntapFileSystem_storageCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOntapFileSystemBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1024"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccOntapFileSystemStorageCapacityConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "2048"),
				),
			},
		},
	})
}

func testAccCheckFsxOntapFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if filesystem == nil {
			return fmt.Errorf("FSx ONTAP File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckFsxOntapFileSystemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_ontap_file_system" {
			continue
		}

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx ONTAP File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckFsxOntapFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx ONTAP File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckFsxOntapFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx ONTAP File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccOntapFileSystemBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccOntapFileSystemSingleAzConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccOntapFileSystemFsxAdminPasswordConfig(rName, pass string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
  fsx_admin_password  = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, pass))
}

func testAccOntapFileSystemEndpointIPAddressRangeConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity          = 1024
  subnet_ids                = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type           = "MULTI_AZ_1"
  throughput_capacity       = 128
  preferred_subnet_id       = aws_subnet.test1.id
  endpoint_ip_address_range = "198.19.255.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemDiskIopsConfigurationConfig(rName string, iops int) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id

  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, iops))
}

func testAccOntapFileSystemRouteTableConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
  route_table_ids     = [aws_route_table.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemSecurityGroupIds1Config(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id]
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemSecurityGroupIds2Config(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccOntapFileSystemTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOntapFileSystemWeeklyMaintenanceStartTimeConfig(rName, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity              = 1024
  subnet_ids                    = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type               = "MULTI_AZ_1"
  throughput_capacity           = 128
  preferred_subnet_id           = aws_subnet.test1.id
  weekly_maintenance_start_time = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, weeklyMaintenanceStartTime))
}

func testAccOntapFileSystemDailyAutomaticBackupStartTimeConfig(rName, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                  = 1024
  subnet_ids                        = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type                   = "MULTI_AZ_1"
  throughput_capacity               = 128
  preferred_subnet_id               = aws_subnet.test1.id
  daily_automatic_backup_start_time = %[2]q
  automatic_backup_retention_days   = 1

  tags = {
    Name = %[1]q
  }
}
`, rName, dailyAutomaticBackupStartTime))
}

func testAccOntapFileSystemAutomaticBackupRetentionDaysConfig(rName string, retention int) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 1024
  subnet_ids                      = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type                 = "MULTI_AZ_1"
  throughput_capacity             = 128
  preferred_subnet_id             = aws_subnet.test1.id
  automatic_backup_retention_days = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, retention))
}

func testAccOntapFileSystemKMSKeyIDConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
  kms_key_id          = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOntapFileSystemThroughputCapacityConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 256
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccOntapFileSystemStorageCapacityConfig(rName string) string {
	return acctest.ConfigCompose(testAccOntapFileSystemBaseConfig(rName), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 2048
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}
