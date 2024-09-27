package hcl2

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tmccombs/hcl2json/convert"
)

const inputa = `
resource "aws_elastic_beanstalk_environment" "example" {
	name        = "test_environment"
	application = "testing"

	setting {
	  namespace = "aws:autoscaling:asg"
	  name      = "MinSize"
	  value     = "1"
	}

	dynamic "setting" {
	  for_each = data.consul_key_prefix.environment.var
	  content {
		heredoc = <<-EOF
		This is a heredoc template.
		It references ${local.other.3}
		EOF
		simple = "${4 - 2}"
		cond = test3 > 2 ? 1: 0
		heredoc2 = <<EOF
			Another heredoc, that
			doesn't remove indentation
			${local.other.3}
			%{if true ? false : true}"gotcha"\n%{else}4%{endif}
		EOF
		loop = "This has a for loop: %{for x in local.arr}x,%{endfor}"
		namespace = "aws:elasticbeanstalk:application:environment"
		name      = setting.key
		value     = setting.value
	  }
	}
  }`

const outputa = `{
    "resource": {
        "aws_elastic_beanstalk_environment": {
            "example": [
                {
                    "application": "testing",
                    "dynamic": {
                        "setting": [
                            {
                                "content": [
                                    {
                                        "cond": "${test3 \u003e 2 ? 1: 0}",
                                        "heredoc": "This is a heredoc template.\nIt references ${local.other.3}\n",
                                        "heredoc2": "\t\t\tAnother heredoc, that\n\t\t\tdoesn't remove indentation\n\t\t\t${local.other.3}\n\t\t\t%{if true ? false : true}\"gotcha\"\\n%{else}4%{endif}\n",
                                        "loop": "This has a for loop: %{for x in local.arr}x,%{endfor}",
                                        "name": "${setting.key}",
                                        "namespace": "aws:elasticbeanstalk:application:environment",
                                        "simple": "${4 - 2}",
                                        "value": "${setting.value}"
                                    }
                                ],
                                "for_each": "${data.consul_key_prefix.environment.var}"
                            }
                        ]
                    },
                    "name": "test_environment",
                    "setting": [
                        {
                            "name": "MinSize",
                            "namespace": "aws:autoscaling:asg",
                            "value": "1"
                        }
                    ]
                }
            ]
        }
    }
}`

const inputb = `
provider "aws" {
    version             = "=2.46.0"
    alias                  = "one"
}
`

const outputb = `{
    "provider": {
        "aws": [
            {
                "alias": "one",
                "version": "=2.46.0"
            }
        ]
    }
}`

const inputc = `
provider "aws" {
    version             = "=2.46.0"
    alias                  = "one"
}
provider "aws" {
    version             = "=2.47.0"
    alias                  = "two"
}
`

const outputc = `{
    "provider": {
        "aws": [
            {
                "alias": "one",
                "version": "=2.46.0"
            },
            {
                "alias": "two",
                "version": "=2.47.0"
            }
        ]
    }
}`

const inputd = `resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = "sub1"
    allocation_id = "eip1"
  }
}`

const outputd = `{
    "resource": {
        "aws_lb": {
            "example": [
                {
                    "load_balancer_type": "network",
                    "name": "example",
                    "subnet_mapping": [
                        {
                            "allocation_id": "eip1",
                            "subnet_id": "sub1"
                        }
                    ]
                }
            ]
        }
    }
}`

const inpute = `resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = "sub1"
    allocation_id = "eip1"
  }

  subnet_mapping {
    subnet_id     = "sub2"
    allocation_id = "eip2"
  }
}`

const outpute = `{
    "resource": {
        "aws_lb": {
            "example": [
                {
                    "load_balancer_type": "network",
                    "name": "example",
                    "subnet_mapping": [
                        {
                            "allocation_id": "eip1",
                            "subnet_id": "sub1"
                        },
                        {
                            "allocation_id": "eip2",
                            "subnet_id": "sub2"
                        }
                    ]
                }
            ]
        }
    }
}`

func TestConversion(t *testing.T) {
	testTable := map[string]struct {
		input  string
		output string
	}{
		"simple-resources": {input: inputa, output: outputa},
		"single-provider":  {input: inputb, output: outputb},
		"two-providers":    {input: inputc, output: outputc},
		"one-block":        {input: inputd, output: outputd},
		"two-blocks":       {input: inpute, output: outpute},
	}

	for name, tc := range testTable {
		testInput := []byte(tc.input)

		convertedInput, err := convert.Bytes(testInput, "", convert.Options{})
		if err != nil {
			t.Fatal("convert bytes:", err)
		}

		var indented bytes.Buffer
		if err := json.Indent(&indented, convertedInput, "", "    "); err != nil {
			t.Fatal("Failed to indent file:", err)
		}

		computedJSON := indented.String()
		if computedJSON != tc.output {
			t.Errorf("For test %s\nExpected:\n%s\n\nGot:\n%s", name, tc.output, computedJSON)
		}
	}
}
