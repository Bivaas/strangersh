@echo off
title strangersh
ssh -o StrictHostKeyChecking=accept-new -p 32128 bore.pub
pause