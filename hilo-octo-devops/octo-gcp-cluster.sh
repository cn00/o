#!/bin/bash
if [ $# -ne 7 ]; then
  echo "usages $./octo-gcp-cluster.sh NETWORKNAME INSTANCE_PREFIX MACNINE_TYPE DB_MACHINE_TYPE ZONE ADD_ZONE REGION" 1>&2
  echo "ex) $./octo-gcp-cluster.sh temp-octo-n test-octo-app-001 n1-standard-1 n1-highmem-2 asia-northeast1-a asia-northeast1-c asia-northeast1" 1>&2
  exit 1
fi

networkname=$1
instancePrefix=$2
machineType=$3
dbMachineType=$4
zone=$5
addzone=$6
region=$7

echo ============= START OCTO CLUSTER =============

# CREATE NETWORK
echo ----------- Start Create Network -----------
gcloud compute networks create "$networkname" --subnet-mode=auto
echo ----------- End Create Network -----------

# CREATE CLUSTERS
echo ----------- Start Create Clusters -----------
gcloud container clusters create "$instancePrefix" --num-nodes 2 --machine-type $machineType --network $networkname --zone $zone --additional-zones $addzone
echo ----------- End Create Clusters -----------
# CREATE MySQL VM
echo ----------- Start Create  MySQL VM -----------
gcloud compute instances create $instancePrefix-mysql --machine-type $dbMachineType --network $networkname --zone $zone --image-family=centos-6 --image-project=centos-cloud
echo ----------- End Create  MySQL VM -----------

# CREATE FIREWALL
echo ----------- Start Create Firewall -----------
gcloud compute firewall-rules create $instancePrefix-allow-http --network $networkname --allow tcp:80 --source-ranges 0.0.0.0/0 --target-tags=http-server --direction=ingress

gcloud compute firewall-rules create $instancePrefix-allow-https --network $networkname --allow tcp:443 --source-ranges 0.0.0.0/0 --target-tags=https-server --direction=ingress

IPRANGE=$(gcloud compute networks subnets describe $networkname --region=$region --format="value(ipCidrRange)")

gcloud compute firewall-rules create $instancePrefix-allow-internal-$zone --network $networkname --allow tcp,icmp,udp --source-ranges $IPRANGE --direction=ingress

gcloud compute firewall-rules create $instancePrefix-allow-ssh --network $networkname --allow tcp:22 --source-ranges 0.0.0.0/0  --direction=ingress

TAGS_ARRAY=()

for instanceInfo in $(
    gcloud compute instances list --filter=name:gke-$instancePrefix --format="csv[no-heading](name,zone)"
    )
do
      IFS=',' read -r -a instanceInfoArray<<< "$instanceInfo"

      NAME="${instanceInfoArray[0]}"
      ZONE="${instanceInfoArray[1]}"
      echo InstanceName:$NAME, Zone:$ZONE
      gcloud config set compute/zone $ZONE
      for instanceTags in $(gcloud compute instances describe $NAME --format="value(tags.items)")
      do
        IFS=',' read -r -a instanceTagsArray<<< "$instanceTags"
        TAGS="${instanceTagsArray[0]}"
        TAGS_ARRAY+=($TAGS)
      done
done
TAGLIST=$(IFS=, ; echo "${TAGS_ARRAY[*]}")
echo Instance Tag List: $TAGLIST
gcloud compute firewall-rules create $instancePrefix-api --network $networkname --allow tcp:30420 --source-ranges 0.0.0.0/0 --target-tags=$TAGLIST --direction=ingress

echo ----------- End Create Firewall -----------

# CREATE LB

echo ----------- Start Create LB -----------
gcloud compute health-checks create tcp $instancePrefix-health-check --port 30420

gcloud compute backend-services create $instancePrefix-backend-service --protocol HTTP --health-checks $instancePrefix-health-check --global --enable-cdn

for instanceGrpInfo in $(
  gcloud compute instance-groups list --filter=name:gke-$instancePrefix --format="csv[no-heading](name,LOCATION)"
  )
do
    IFS=',' read -r -a instanceGrpInfoArray<<< "$instanceGrpInfo"
    GRPNAME="${instanceGrpInfoArray[0]}"
    GRPLOCATION="${instanceGrpInfoArray[1]}"
    echo Instance Group Name: $GRPNAME, Instance Group Zone : $GRPLOCATION
    gcloud compute backend-services add-backend $instancePrefix-backend-service --balancing-mode UTILIZATION --max-utilization 0.8 --capacity-scaler 1 --instance-group $GRPNAME --instance-group-zone $GRPLOCATION --global
    gcloud compute instance-groups managed set-named-ports "$GRPNAME" --named-ports "http:30420" --zone "$GRPLOCATION"
done

gcloud compute url-maps create $instancePrefix-lb --default-service $instancePrefix-backend-service
gcloud compute target-http-proxies create $instancePrefix-http-lb-proxy --url-map $instancePrefix-lb
gcloud compute forwarding-rules create $instancePrefix-http-lb-forwading-rule --global --target-http-proxy $instancePrefix-http-lb-proxy --ports 80 --global-address
echo ----------- End Create LB -----------

echo ============= END OCTO CLUSTER =============
