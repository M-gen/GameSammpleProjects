using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class CubeController : MonoBehaviour
{
    private enum KeyResult
    {
        None,
        Down,
        Up,
        Left,
        Right,
        GameExit,
    }

    KeyResult keyResult = KeyResult.None;

    const float AnimationTimerMax = 0.8f;
    const float MoveTimerMax = 0.3f;

    Vector3 nextPosition;
    Vector3 prePosition;
    float animationTimer = 0;
    float moveTimer = 0;

    void Start()
    {
        prePosition = nextPosition = transform.position;
    }

    private void UpdateInput()
    {
        keyResult = KeyResult.None;

        if (Input.GetKey(KeyCode.W))
        {
            keyResult = KeyResult.Up;
        }
        else if (Input.GetKey(KeyCode.S))
        {
            keyResult = KeyResult.Down;
        }
        else if (Input.GetKey(KeyCode.A))
        {
            keyResult = KeyResult.Left;
        }
        else if (Input.GetKey(KeyCode.D))
        {
            keyResult = KeyResult.Right;
        }
        else if (Input.GetKey(KeyCode.Escape))
        {
            keyResult = KeyResult.GameExit;
        }
    }

    private void UpdateDatas()
    {
        switch (keyResult)
        {
            case KeyResult.Up:
                if (moveTimer == 0)
                {
                    nextPosition += new Vector3(0, 1, 0);
                    moveTimer = MoveTimerMax;
                }
                break;
            case KeyResult.Down:
                if (moveTimer==0)
                {
                    nextPosition += new Vector3(0,-1,0);
                    moveTimer = MoveTimerMax;
                }
                break;
            case KeyResult.Left:
                if (moveTimer == 0)
                {
                    nextPosition += new Vector3(-1, 0, 0);
                    moveTimer = MoveTimerMax;
                }
                break;
            case KeyResult.Right:
                if (moveTimer == 0)
                {
                    nextPosition += new Vector3(1, 0, 0);
                    moveTimer = MoveTimerMax;
                }
                break;
            case KeyResult.GameExit:
                Application.Quit();
                break;
        }

        if(moveTimer!= 0)
        {
            moveTimer -= Time.deltaTime;
            if( moveTimer <= 0)
            {
                moveTimer = 0;
                transform.position = prePosition = nextPosition;
            }
            else
            {
                var v = nextPosition - prePosition;
                transform.position = prePosition + v * (MoveTimerMax - moveTimer) / MoveTimerMax;
            }
        }

        animationTimer+=Time.deltaTime;
        if (animationTimer >= AnimationTimerMax)
        {
            animationTimer -= AnimationTimerMax;
        }
        gameObject.GetComponent<Renderer>().material.color = new Color(1.0f, 0.5f + animationTimer/AnimationTimerMax * 0.3f, 0.5f + animationTimer / AnimationTimerMax * 0.3f);
    }

    void Update()
    {
        UpdateInput();
        UpdateDatas();
    }
}
