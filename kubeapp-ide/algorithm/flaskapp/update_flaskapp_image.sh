echo "build flask app image"
docker build . -t algo-sentiment:0.1
cd ~
echo "save image"
docker save algo-sentiment:0.1 > algo-sentiment_0.1.tar
echo "transport image to node"
scp ./algo-sentiment_0.1.tar root@slave1:~/images/
scp ./algo-sentiment_0.1.tar root@slave2:~/images/
ssh -t -p 22 root@slave1 docker load < ~/images/algo-sentiment_0.1.tar
ssh -t -p 22 root@slave2 docker load < ~/images/algo-sentiment_0.1.tar
